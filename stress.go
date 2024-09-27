package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/common-nighthawk/go-figure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	stressTestDuration time.Duration
	stressTestMutex    sync.Mutex
	stressTestRunning  bool
)

type Config struct {
	APIKey        string         `json:"apiKey"`
	DonationTiers []DonationTier `json:"donationTiers"`
}

type DonationTier struct {
	Amount  float64 `json:"amount"`
	Seconds int     `json:"seconds"`
}

var config Config

func init() {
	// Carica la configurazione
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Errore durante la lettura del file di configurazione: %v", err)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Errore durante l'analisi del file di configurazione: %v", err)
	}

	// Ordina i livelli di donazione in ordine decrescente per importo
	sort.Slice(config.DonationTiers, func(i, j int) bool {
		return config.DonationTiers[i].Amount > config.DonationTiers[j].Amount
	})
}

func printASCIIArt() {
	myFigure := figure.NewFigure("QDSStress", "", true)
	myFigure.Print()
}

func main() {
	printASCIIArt()
	log.Println("Avvio del programma...")

	// Avvia la goroutine che gestisce l'accodamento delle richieste
	go handleSuperChatDonations()

	log.Println("Programma avviato. In attesa di donazioni Super Chat...")

	// Mantieni la goroutine principale in esecuzione
	select {}
}

func runStressTest(seconds int) {
	stressTestMutex.Lock()
	if stressTestRunning {
		stressTestDuration += time.Duration(seconds) * time.Second
		stressTestMutex.Unlock()
		return
	}
	stressTestRunning = true
	stressTestDuration = time.Duration(seconds) * time.Second
	stressTestMutex.Unlock()

	go func() {
		startTime := time.Now()
		fmt.Printf("Avvio del test di stress estremo per %d secondi\n", seconds)

		// Stress della CPU - utilizzo al 100%
		cpuCount := runtime.NumCPU()
		for i := 0; i < cpuCount; i++ {
			go func() {
				for time.Since(startTime) < stressTestDuration {
					// Ciclo stretto per massimizzare l'uso della CPU
					for j := 0; j < 1000000; j++ {
						_ = rand.Int()
					}
				}
			}()
		}

		// Stress della RAM - consuma tutta la memoria disponibile
		var memoryHog [][]byte
		for time.Since(startTime) < stressTestDuration {
			// Alloca 1GB alla volta
			chunk := make([]byte, 1024*1024*1024)
			for i := range chunk {
				chunk[i] = byte(rand.Intn(256))
			}
			memoryHog = append(memoryHog, chunk)
			runtime.GC() // Forza la garbage collection per assicurarsi di utilizzare tutta la memoria disponibile
		}

		// Mantieni la memoria allocata
		runtime.KeepAlive(memoryHog)

		stressTestMutex.Lock()
		stressTestRunning = false
		stressTestMutex.Unlock()
		fmt.Println("Test di stress estremo completato (se il sistema è ancora reattivo)")
	}()
}

func handleSuperChatDonations() {
	ctx := context.Background()

	log.Println("Lettura delle credenziali OAuth2...")
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Impossibile leggere il file delle credenziali client: %v", err)
	}

	log.Println("Configurazione del flusso OAuth2...")
	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		log.Fatalf("Impossibile analizzare le credenziali del client: %v", err)
	}

	log.Println("Ottenimento del client HTTP autenticato...")
	client := getClient(config)

	log.Println("Creazione del servizio YouTube...")
	youtubeService, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Errore durante la creazione del client YouTube: %v", err)
	}

	log.Println("Connesso e in ascolto per eventi Super Chat...")

	processedEvents := make(map[string]bool)

	for {
		log.Println("Recupero degli eventi Super Chat...")
		call := youtubeService.SuperChatEvents.List([]string{"id", "snippet"}).
			MaxResults(50)

		response, err := call.Do()
		if err != nil {
			log.Printf("Errore durante il recupero degli eventi Super Chat: %v", err)
			log.Println("Attesa di 10 secondi prima di riprovare...")
			time.Sleep(10 * time.Second)
			continue
		}

		log.Printf("Recuperati %d eventi Super Chat", len(response.Items))

		for _, item := range response.Items {
			if !processedEvents[item.Id] {
				amount := float64(item.Snippet.AmountMicros) / 1000000.0
				currency := item.Snippet.Currency
				displayName := item.Snippet.SupporterDetails.DisplayName

				log.Printf("Ricevuto Super Chat: %s ha donato %.2f %s", displayName, amount, currency)

				donationSeconds := getDonationSeconds(amount)

				log.Printf("Aggiunta di %d secondi al test di stress", donationSeconds)
				runStressTest(donationSeconds)

				processedEvents[item.Id] = true
			}
		}

		// Attendi 10 secondi prima di controllare nuovamente
		log.Println("Attesa di 10 secondi prima del prossimo controllo...")
		time.Sleep(10 * time.Second)
	}
}

func getDonationSeconds(amount float64) int {
	for _, tier := range config.DonationTiers {
		if amount >= tier.Amount {
			return tier.Seconds
		}
	}
	return config.DonationTiers[len(config.DonationTiers)-1].Seconds
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		log.Println("Token non trovato o non valido. Richiesta di un nuovo token...")
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	} else {
		log.Println("Token esistente caricato con successo.")
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Vai a questo link nel tuo browser e autorizza l'applicazione:\n%v\n", authURL)
	log.Println("Inserisci il codice di autorizzazione:")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Impossibile leggere il codice di autorizzazione: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Impossibile recuperare il token dal web: %v", err)
	}
	log.Println("Nuovo token ottenuto con successo.")
	return tok
}

// tokenFromFile recupera un token da un file locale.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.Unmarshal(f, tok)
	return tok, err
}

// saveToken salva il token in un file.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Salvataggio del token nel file: %s\n", path)
	f, err := json.Marshal(token)
	if err != nil {
		log.Fatalf("Impossibile codificare il token in JSON: %v", err)
	}
	if err := ioutil.WriteFile(path, f, 0600); err != nil {
		log.Fatalf("Impossibile scrivere il token nel file: %v", err)
	}
}

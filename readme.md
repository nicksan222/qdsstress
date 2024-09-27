## Guida all'Installazione e Utilizzo del Programma di Monitoraggio Super Chat

### Indice
1. Installazione di Go (Golang)
2. Configurazione dell'Ambiente di Sviluppo
Creazione del Progetto
Configurazione del Progetto Google Cloud
Preparazione del Codice
Esecuzione del Programma
Risoluzione dei Problemi Comuni

### 1. Installazione di Go (Golang)
1. Visita il sito ufficiale di Go: https://golang.org/dl/
2. Scarica l'installer per Windows
Esegui l'installer e segui le istruzioni sullo schermo
4. Al termine, apri il Prompt dei Comandi e digita go version per verificare l'installazione


### Clonare il progetto

```
git clone https://github.com/nicksan222/qdsstress.git
```


### Configurazione del Progetto Google Cloud

Vai alla Console Google Cloud: https://console.cloud.google.com/
Crea un nuovo progetto se richiesto, loggando con l'account Google associato al canale YouTube
Vai su API e servizi > Credenziali
Premere il pulsante + ABILITA API E SERVIZI
Abilita l'API di YouTube Data v3
Premere il pulsante CREDENZIALI
Crea le credenziali OAuth 2.0
Scarica il file JSON delle credenziali e rinominalo client_secret.json
Sposta client_secret.json nella cartella del progetto

### File di configurazione

Puoi modificare il file config.json per modificare i seguenti parametri:

```
{
  "donationTiers": [
    { "amount": 1, "seconds": 1 },
    { "amount": 5, "seconds": 5 },
    { "amount": 10, "seconds": 15 },
    { "amount": 20, "seconds": 30 },
    ecc...
  ]
}
```

Il file dice quanto deve essere aumentato il credito del bot per ogni donazione (prenderà sempra il più vicino per difetto)

### Installare i pacchetti

La prima volta che esegui il programma dovrai installare i pacchetti necessari
Lancia il comando:
```
go mod tidy
```

### Esecuzione del programma

Per eseguire il programma, apri il terminale nella cartella del progetto e digita il seguente comando:

```
go run stress.go
```


Verrà mostrato il link di autorizzazione, clicca sul link e accedi con il tuo account YouTube.
Verrai reindirizzato a un link che darà errore, non preoccuparti, il programma funziona comunque.
Il link sarà qualcosa del genere:

```
http://localhost/?state=random-string&code=4/0AQlEd8x1geTHDcWN4qNE00uYOCvuvB0Fd5QP7_Jp_dQPlV5AbXmqTtuZIIULFcQIp8mbHg&scope=https://www.googleapis.com/auth/youtube
```

Copia da dopo "code=" e fino a &scope= e incolla nel programma prima di startare il test
Nell'esempio sopra avrai bisogno di copiare e incollare:

```
4/0AQlEd8x1geTHDcWN4qNE00uYOCvuvB0Fd5QP7_Jp_dQPlV5AbXmqTtuZIIULFcQIp8mbHg
```

Incolla nel programma e premi invio, il programma partirà e ascolterà i Super Chat del tuo canale


## Limitazioni
- Il programma può monitorare solo un canale YouTube alla volta.
- Le donazioni inferiori all'importo minimo configurato non attiveranno il test di stress.
# qdsstress

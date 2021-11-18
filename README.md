# Multicast totalmente e causalmente ordinato in Go

Lo scopo del progetto è quello di realizzare, nel linguaggio di programmazione Go, un'applicazione distribuita che implementi gli algoritmi di multicast totalmente ordinato e causalmente ordinato.

L'applicazione prevede:
- un servizio di registrazione dei processi che partecipano al gruppo di comunicazione multicast;
- il supporto dei seguenti algoritmi di multicast:
  - multicast totalmente ordinato implementato in modo centralizzato tramite un sequencer;
  - multicast totalmente ordinato implementato in modo decentralizzato tramite l'uso di clock logici scalari;
  - multicast causalmente ordinato implementato in modo decentralizzato tramite l'uso di clock logici vettoriali.

## Quickstart

1. **Clona questo repository:**

```
git clone https://github.com/g3lbin/sdcc-project.git
cd sdcc-project
```

2. [Opzionale] **Modifica i parametri che definiscono l'ambiente dei servizi:**

Il file [`.env`](deployments/.env) presenta già una configurazione di default:
```
## Default network configuration
REGISTRATION_PORT=4321
SEQUENCER_PORT=5555
MULTICAST_PORT=1234
CONTROL_PORT=8888

## Application's parameters
MEMBERS_NUM=4
DELAY=5
## Uncomment only one of the following lines
MULTI_ALGO=centralized_totally_ordered_multicast
# MULTI_ALGO=decentralized_totally_ordered_multicast
# MULTI_ALGO=causally_ordered_multicast

```
È possibile cambiare il valore di tutti i parametri ad eccezione di `MULTI_ALGO`. Per specificare l'algoritmo di multicast che si intende utilizzare, è sufficiente lasciare *uncommented* una sola delle tre linee corrispondenti a tale parametro.

> **Nota:** L'ultima riga del file deve essere una *blank line*.

3. **Esegui la build dei servizi e avviali:**

```
./scripts/compose.sh up
```

4. **Interagisci con il gruppo di comunicazione:**
```
go run ./cmd/main.go [-v]
```
Specificando il flag `-v` si esegue il programma in modalità *verbose*.

5. [Opzionale] **Clean up:**
```
./scripts/compose.sh down
```

## Testing
È possibile testare il funzionamento degli algoritmi di multicast nel caso vi sia un solo peer ad inviare messaggi e nel caso in cui ve ne siano molteplici.

Per eseguire ciascun caso di test è necessario:
1. ripercorrere i passi `2` e `3` descritti in `Quickstart` 
2. lanciare il comando:
  ```
  go run ./test/test.go -algorithm <num> -source <type>
  ```
  - `num` può assumere valore `1`, `2` o `3` a seconda dell'algoritmo specificato in `[.env](deployments/.env)`:

    - `1` (centralized_totally_ordered_multicast)
    - `2` (decentralized_totally_ordered_multicast)
    - `3` (causally_ordered_multicast)
  - `source` può assumere valore `single` o `multiple` per specificare se uno o più peer possono inviare messaggi.

## Documentazione
Per avere maggiori informazioni riguardo il progetto, è possibile vedere la [relazione](docs/relazione.pdf) che ne riporta tutti i dettagli.

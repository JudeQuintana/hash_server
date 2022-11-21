# Hash Server 9000

```
     ____.             ________        ________
    |    |____  ___.__.\_____  \       \_____  \   ____   ____
    |    \__  \<   |  | /  / \  \       /   |   \ /    \_/ __ \
/\__|    |/ __ \\___  |/   \_/.  \     /    |    \   |  \  ___/
\________(____  / ____|\_____\ \_/_____\_______  /___|  /\___  >
              \/\/            \__>_____/       \/     \/     \/

--=[ PrEsENtZ ]=--

--=[ HasH SeRVeR 9000 ]=--

--=[ #HoLLaAtYaBoi ]=--
```
## Moar Idiomatic

This is my attempt at learning idomatic Golang by building a non-persistent password hashing service using only the standard library.  My biggest challenges were understanding slices, channels, goroutines, concurrency with shared variables and testing. This service is meant for learning, not for use in production.

Hash Server 9000 is able to process multiple connections simultaneously (concurrency safe).

Run Server: `go run main.go`

Run Queries while server is servin': `./query.sh`

Run Tests: `cd hasher/`, `go test`

Build Server: `go build -o hash_server`

Run Server Executable: `./hash_server`

#### Endpoints

1. `POST /hash` - Hash and encode a password string. The request must contain a `password` parameter. Returns the `id` of Base64 encoded string of the password that's been hashed with SHA512 with a 5 second delay to simulate asynchronous processing. Example: `curl --data "password=angryMonkey" http://localhost:8080/hash`

2. `GET /hash/:id` - Retrieve a generated hash with the `id` (after approximately 5 seconds), otherwise you will receive error `id not found`.  Example: `curl http://localhost:8080/hash/1` will return: `ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==`.

3. `GET /stats` - Statistics endpoint. Returns a JSON object with the `total` count of the number of password hash requests made to the server so far and the `average` time in milliseconds it has taken to process all of the requests.  Example: `curl http://localhost:8080/stats`, will return: `{"total": 1, "average": 123}`

4. `GET /shutdown` - Graceful shutdown. When a request is made to `/shutdown` the server will reject new requests and wait for any pending/in-flight work to finish before exiting.


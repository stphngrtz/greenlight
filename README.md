# Greenlight
This projects covers the topics from the book [Let's go further](https://lets-go-further.alexedwards.net/) by [Alex Edwards](https://github.com/alexedwards).

Run the Postgres database instance with Docker.
```bash
docker run --name pg -p 5432:5432 -e POSTGRES_USER=user -e POSTGRES_PASSWORD=my-secret-pw -e POSTGRES_DB=greenlight -d postgres:16
```

Connect to the Postgres database instance and login as user.
```bash
docker exec -it pg bash
psql --dbname=greenlight --username=user
```

Run the application.
```bash
go run ./cmd/api
```

Send requests to the application.
```bash
# valid request
curl -iX POST -d '{"title":"Moana","year":2016,"runtime":"107 mins","genres":["animation","adventure"]}' localhost:4000/v1/movies
# invalid request
curl -iX POST -d '{"title":"","year":1000,"runtime":"-123 mins","genres":["sci-fi","sci-fi"]}' localhost:4000/v1/movies
```

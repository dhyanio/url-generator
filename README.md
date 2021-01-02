# url-generator

## A small short URL Generator in golang

This is a small short URL generator writen in Golang. There are majorly 3 technologies I have used here.

1. Golang : Latest
2. PostreSQL : 13
3. Redis : 6.0

PostgreSQL for the database and Redis for the cashing. Cashing help us to get good performance and low latancy in our program.
I will be using <b>GORM</b> library for connecting to the DB and <b>Go-Redis</b> for the Redis cache.

### Step 1

Use these command to download them for our Golang program environment.
```bash
go get -u github.com/jinzhu/gorm
go get -u github.com/go-redis/redis

```

### Step 2
I can establish a connection to our PostgreSQL database using the below connection string
```bash
postgres", "host=127.0.0.1 port=5432 user=postgres dbname=tiny_scale_go password=<db password> sslmode=disable

```
### Step 3
We provide the Redis server config as per the below snippet
```bash
&redis.Options{
Addr:     "localhost:6379",
Password: "",
DB:       0,
}
```

Improvements are welcome! HERE

Ajay K. Dhyani < dhyani.devops@gmail.com >

Golang developer/DevOps engineer

Gurugram, India
https://medium.com/@hpmahesh73/creating-a-simple-tiny-url-generator-using-golang-postgresql-and-redis-df8a29f2deab

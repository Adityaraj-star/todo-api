# Todo API (Go)

This is a basic backend project I built while learning Go.
The idea was to understand how REST APIs work and how to structure a backend project properly instead of writing everything in one file.

It’s a simple Todo API where you can create, update, delete and fetch tasks.

---

## What it does

* Add a new todo
* Get all todos
* Get a single todo using id
* Update a todo
* Delete a todo
* Logs every request (method + path + time taken)

---

## Tech used

* Go (net/http)
* JSON (encoding/json)
* sync.Mutex (for safe concurrent access)
* UUID for unique IDs

---

## Project structure (rough idea)

```
handler/     -> handles HTTP requests
store/       -> stores data (in-memory for now)
model/       -> defines Todo struct
middleware/  -> logging middleware
main.go      -> starts the server
```

I tried to keep things separate so it feels closer to how real backend projects are structured.

---

## How to run

Clone the repo:

```
git clone https://github.com/your-username/todo-api.git
cd todo-api
```

Run:

```
go run main.go
```

Server runs on:

```
http://localhost:8080
```

---

## API routes

**GET /todos**
returns all todos

**GET /todos/{id}**
returns a single todo

**POST /todos**
create a todo
body:

```
{
  "title": "learn go"
}
```

**PUT /todos/{id}**
update a todo
body:

```
{
  "title": "learn go properly",
  "status": "in-progress"
}
```

**DELETE /todos/{id}**
delete a todo

---

## Things I learned

* How to build APIs using Go without any framework
* How routing works with `http.ServeMux`
* Structuring code into different layers (handler, store, etc.)
* Handling JSON requests/responses
* Basic concurrency using mutex
* Writing simple middleware

---

## Limitations

* Data is stored in memory (resets when server restarts)
* No database yet
* No authentication
* Basic validation

---

## What I plan to improve

* Add a database (PostgreSQL or MongoDB)
* Better validation
* Maybe use a router like chi or gorilla/mux
* Add some extra features like filtering or pagination

---

## Author

Aditya Raj

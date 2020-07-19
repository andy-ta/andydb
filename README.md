# AndyDB

Born out of necessity during a hackathon, AndyDB is a quick-and-easy REST-based NoSQL database inspired by MongoDB.

AndyDB is still in very early alpha and if you are not in my immediate friend circle then it is not ACID-compliant.

## Usage

You can either build the binary yourself with `go build` or [download](https://github.com/mockoon/mockoon/releases)
the release.

After starting `andydb.exe`, you can make simple RESTful CRUD requests to the server at `http://localhost:42069/api`.

For example, `curl -d '{"email": "andy@andy.db"}' http://localhost:42069/api/contacts` will create the contacts resource
type (since it does not exist yet) and will save the provided body as an entry of that resource. 
It will return with the created object in JSON format, and with a new field `_id`, to be used for future operations.
Subsequent POST requests to the contacts resource will append the entry to the list of contacts.

With the `_id` you may now perform a GET / PUT / DELETE requests in the format of:

- GET `http://localhost:42069/api/contacts/{_id}`
  - `curl http://localhost:42069/api/contacts/{_id}`
- PUT `http://localhost:42069/api/contacts/{_id}` 
  - `curl -X PUT -d '{"email": "db@andy.db"}' http://localhost:42069/api/contacts/{_id}`
- DELETE `http://localhost:42069/api/contacts/{_id}` (
  - `curl -X PUT http://localhost:42069/api/contacts/{_id}`

If you don't provide the id for a GET request, it will return all entries of the resource.

## Upcoming Features

- Locking
- Save to file
- Security
- Containerization
- Type checking
- Cleanups by coding in Go better
- Etc.
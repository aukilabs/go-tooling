# rest

A package that provides basic HTTP mechanisms to build a rest API.

## Usage

```go
func main() {
    h := handler{
        BaseHandler: rest.BaseHandler{
			Encode: json.Marshal,
			Decode: json.Unmarshal,
    }

    api := rest.NewMux()
    api.HandleFunc(http.MethodGet, "/tests", handleGet)
    api.HandleFunc(http.MethodPost, "/tests", handlePost)

    http.ListenAndServe(":8080", api)
}


type handler struct {
    BaseHandler
}

func (h handler) handleGet(w http.ResponseWriter, r *http.Request) {
    h.Ok(w, r)
}

func (h handler) handlePost(w http.ResponseWriter, r *http.Request) {
    h.Ok(w, r, struct{
        Msg string `json:"msg"`
    }{
        Msg: "hello",
    })
}
```

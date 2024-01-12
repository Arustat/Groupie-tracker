package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
)

func getContentByKey(key string) (string, error) {
    apiURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/%s", key)

    // Faire une requête à votre API pour obtenir le contenu pour la clé donnée
    response, err := http.Get(apiURL)
    if err != nil {
        return "", err
    }
    defer response.Body.Close()

    // Lire la réponse de l'API
    apiContent, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return "", err
    }

    return string(apiContent), nil
}

func getAPIContent(w http.ResponseWriter, r *http.Request) {
    key := r.URL.Path[1:] // Récupérer la clé à partir de l'URL
    content, err := getContentByKey(key)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(content))
}

func main() {
    http.Handle("/", http.FileServer(http.Dir("./"))) // Servir les fichiers statiques (index.html)
    http.HandleFunc("/artists", getAPIContent)
    http.HandleFunc("/dates", getAPIContent)
    http.HandleFunc("/locations", getAPIContent)
    
    log.Fatal(http.ListenAndServe(":8010", nil))
}

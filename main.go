package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Intégre le dossier asset dans le serveur
	http.Handle("/asset/", http.StripPrefix("/asset/", http.FileServer(http.Dir("asset"))))

	// Définit la route principale
	http.HandleFunc("/", indexHandler)

	// Définit la route de recherche
	http.HandleFunc("/search", searchHandler)

	// Définit la route des suggestions avec nom artistes
	http.HandleFunc("/suggest", suggestHandler)

	// Définit la route des suggestions avec géocalisation
	http.HandleFunc("/suggestgeo", suggest_geoHandler)

	//Définit la route pour agir en tant que proxy vers l'Api
	http.HandleFunc("/geonames", handleGeonamesProxy)

	// Lance le serveur
	log.Fatal(http.ListenAndServe(":8000", nil))

}

// Structure artist pour pouvoir utiliser les données json de l'api artist
type ArtistsInfo struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
	showConcert  bool
}

// Structure Locations pour pouvoir utiliser les données json de l'api Locations
type LocationsInfo struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
		Dates     string   `json:"dates"`
	} `json:"index"`
}

// Structure Dates pour pouvoir utiliser les données json de l'api Dates
type DatesInfo struct {
	Index []struct {
		ID    int      `json:"id"`
		Dates []string `json:"dates"`
	} `json:"index"`
}

// Structure Relation pour pouvoir utiliser les données json de l'api Relation
type RelationsInfo struct {
	Index []struct {
		ID             int                 `json:"id"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}

// Structure Groupie tracker pour pouvoir utiliser les données json de l'api groupie tracker
type GroupieTracker struct {
	ArtistsInfo   string `json:"artists"`
	LocationsInfo string `json:"locations"`
	DatesInfo     string `json:"dates"`
	RelationsInfo string `json:"relation"`
}

// Structure pour ranger les filtres
type Filters struct {
	Search   string
	Date     string
	Location string
}

func recupJSON() (*GroupieTracker, error) {
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api")
	if err != nil {
		log.Printf("Erreur lors de la requête GET : %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	var apiInfo GroupieTracker
	err = json.NewDecoder(resp.Body).Decode(&apiInfo)
	if err != nil {
		log.Printf("Erreur lors du décodage JSON : %v\n", err)
		return nil, err
	}
	return &apiInfo, nil
}

const htmlTemplatePath = "index.html"

func indexHandler(w http.ResponseWriter, r *http.Request) {
	apiInfo, err := recupJSON()
	if err != nil {
		http.Error(w, "Erreur de récupération des infos API", http.StatusInternalServerError)
		return
	}

	artistList, err := recupArtistes(apiInfo.ArtistsInfo)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		return
	}
	// Créer un nouveau template à partir du fichier HTML
	tmpl, err := template.ParseFiles(htmlTemplatePath)
	if err != nil {
		http.Error(w, "Erreur lors de la création du template", http.StatusInternalServerError)
		return
	}

	// Exécuter le template en passant la liste des artistes
	err = tmpl.Execute(w, artistList)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution du template", http.StatusInternalServerError)
		return
	}
}

func handleGeonamesProxy(w http.ResponseWriter, r *http.Request) {
	// Récupérer les paramètres de la requête
	lat := r.URL.Query().Get("lat")
	lng := r.URL.Query().Get("lng")

	// Faire la requête à l'API Geonames
	url := fmt.Sprintf("http://api.geonames.org/findNearbyPlaceNameJSON?lat=%s&lng=%s&username=maymay", lat, lng)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Erreur lors de la requête à l'API Geonames", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Lire la réponse de l'API Geonames
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur lors de la lecture de la réponse de l'API Geonames", http.StatusInternalServerError)
		return
	}

	// Renvoyer la réponse de l'API Geonames
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer les paramètres de recherche depuis la requête
	search := r.URL.Query().Get("search")
	date := r.URL.Query().Get("filtre")
	location := r.URL.Query().Get("localisation")
	yearstr := r.URL.Query().Get("year")
	membre := r.URL.Query().Get("members")
	first_album := r.URL.Query().Get("first_album")

	var membres int

	if membre == "" {
		membres, _ = strconv.Atoi(membre)
		log.Print(membres)
	}

	var year int

	if yearstr == "" {
		year, _ = strconv.Atoi(yearstr)
		log.Print("Nbr membres : ", year)
	}

	apiInfo, err := recupJSON()
	if err != nil {
		http.Error(w, "Erreur de récupération des infos API", http.StatusInternalServerError)
		return
	}

	artistList, err := recupArtistes(apiInfo.ArtistsInfo)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		return
	}

	// Filtrer les données en fonction du nom de l'artiste
	filterDataBySearch, err := filterDataBySearch(artistList, search)
	if err != nil {
		http.Error(w, "Erreur lors du filtrage des données par recherche", http.StatusInternalServerError)
		return
	}
	var formattedDate string
	var filteredByDate []ArtistsInfo
	if date != "" {
		// Convertir la date dans le bon format si nécessaire
		parsedDate, err := time.Parse("2006-01-02", date) // Utilisez le format JJ-MM-AAAA
		if err != nil {
			http.Error(w, "Format de date invalide", http.StatusBadRequest)
			return
		}
		formattedDate = parsedDate.Format("02-01-2006") // Convertir la date en format AAAA-MM-JJ
		log.Println(formattedDate)

		dateList, err := recupDates(apiInfo.DatesInfo)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des dates", http.StatusInternalServerError)
			return
		}
		filteredByDate, err = filterDataByDate(dateList.Index, formattedDate)
		if err != nil {
			http.Error(w, "Erreur lors du filtrage des données par date", http.StatusInternalServerError)
			return
		}
	}

	// Filtrer les données par emplacement si un emplacement est spécifié
	var filteredDataByLocation []ArtistsInfo
	locationList, err := recupLocation(apiInfo.LocationsInfo)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des emplacements", http.StatusInternalServerError)
		return
	}
	filteredDataByLocation, err = filterDataByLocations(locationList.Index, location)
	if err != nil {
		http.Error(w, "Erreur lors du filtrage des données par emplacement", http.StatusInternalServerError)
		return
	}

	// Filtrer les données par relation si tous les filtres sont remplis
	var Results []ArtistsInfo
	relationList, err := recupRelation(apiInfo.RelationsInfo)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des relations", http.StatusInternalServerError)
		return
	}

	if search != "" && location == "" && date == "" {
		Results = filterDataBySearch
	} else if search == "" && location == "" && date != "" {
		Results = filteredByDate
	} else if search == "" && location != "" && date == "" {
		Results = filteredDataByLocation
	} else if search == "" && location == "" && date == "" {
		Results = artistList
	} else {
		Results, err = filterDataByRelations(relationList.Index, formattedDate, location, search)
		if err != nil {
			http.Error(w, "Erreur lors du filtrage des données par relations", http.StatusInternalServerError)
			return
		}
	}

	// Récupérer les paramètres de recherche depuis la requête
	sortA := r.FormValue("alpha")
	log.Print("Sort alpha value:", sortA)

	//Verifier si la case à cocher "alpha" a été cochée
	if sortA == "on" {
		Results, err = trier_ordre_alphabe(Results)
		if err != nil {
			http.Error(w, "Erreur lors du triage des artistes ", http.StatusInternalServerError)
			return
		}
	}

	concert := r.FormValue("concert")
	log.Print("Concert sort value: ", concert)
	if concert == "on" {
		dateList, err := recupDates(apiInfo.DatesInfo)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des dates", http.StatusInternalServerError)
			return
		}
		Results, err = trier_ordre_concert_récent(dateList.Index)
		if err != nil {
			http.Error(w, "Erreur lors du triage des concerts", http.StatusInternalServerError)
			return
		}
		// Définir ShowConcert sur true pour chaque artiste
		for _, artist := range Results {
			artist.showConcert = true
		}
	}

	log.Print(yearstr)

	if yearstr != "1950" {
		Results, err = filterDataByYear(Results, year)
		if err != nil {
			http.Error(w, "Erreur lors du filtrage par date", http.StatusInternalServerError)
			return
		}
	}

	if membres != 0 {
		Results, err = filterDatabyMembers(Results, membres)
		if err != nil {
			http.Error(w, "Erreur lors du filtrage par date", http.StatusInternalServerError)
			return
		}
	}

	if first_album != "" {
		parsedDate, err := time.Parse("2006-01-02", first_album) // Utilisez le format JJ-MM-AAAA
		if err != nil {
			http.Error(w, "Format de date invalide", http.StatusBadRequest)
			return
		}
		f_first_album := parsedDate.Format("02-01-2006") // Convertir la date en format AAAA-MM-JJ
		log.Println("First album : ", f_first_album)

		Results, err = filterDatabyFirstAlbum(Results, f_first_album)
		if err != nil {
			http.Error(w, "Erreur lors du filtrage par date", http.StatusInternalServerError)
			return
		}
	}

	// Exécuter le template en passant les résultats filtrés
	tmpl, err := template.ParseFiles(htmlTemplatePath)
	if err != nil {
		http.Error(w, "Erreur lors de la création du template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, Results)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution du template", http.StatusInternalServerError)
		return
	}
}

func suggestHandler(w http.ResponseWriter, r *http.Request) {
	suggest := r.URL.Query().Get("query")

	apiInfo, err := recupJSON()
	if err != nil {
		http.Error(w, "Erreur de récupération des infos API", http.StatusInternalServerError)
		return
	}

	artistList, err := recupArtistes(apiInfo.ArtistsInfo)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		return
	}

	var suggestions []string

	// Parcourir tous les artistes et vérifier s'ils correspondent à l'entrée de l'utilisateur
	for _, artist := range artistList {
		// Vérifier si le nom de l'artiste correspond à la suggestion de l'utilisateur
		if strings.Contains(strings.ToLower(artist.Name), strings.ToLower(suggest)) {
			suggestions = append(suggestions, artist.Name+" - Artiste(s)")
		}

		// Vérifier tous les membres du groupe
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), strings.ToLower(suggest)) {
				suggestions = append(suggestions, member+" - Membre")
				log.Println(suggestions)

			}
		}
	}

	json.NewEncoder(w).Encode(suggestions)
}
func suggest_geoHandler(w http.ResponseWriter, r *http.Request) {
	suggest := r.URL.Query().Get("query")

	apiInfo, err := recupJSON()
	if err != nil {
		http.Error(w, "Erreur de récupération des infos API", http.StatusInternalServerError)
		return
	}

	geoList, err := recupLocation(apiInfo.LocationsInfo)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		return
	}

	var suggestions []string

	// Parcourir tous les emplacements et vérifier s'ils correspondent à l'entrée de l'utilisateur
	for _, index := range geoList.Index {
		for _, location := range index.Locations {
			locationavecspace := strings.ReplaceAll(location, "_", " ")
			if strings.Contains(strings.ToLower(locationavecspace), strings.ToLower(suggest)) {
				ville, pays := recupvilleetpays(location, suggestions)
				if ville != "" && pays != "" {
					suggestions = append(suggestions, ville, pays)
				} else if ville != "" && pays == "" {
					suggestions = append(suggestions, ville)
				} else if pays != "" && ville == "" {
					suggestions = append(suggestions, pays)
				}
			}
		}
	}

	log.Println(suggestions)

	json.NewEncoder(w).Encode(suggestions)
}

func recupvilleetpays(location string, suggestions []string) (string, string) {
	// Vérifier si l'emplacement contient un tiret ("-")
	if strings.Contains(location, "-") {
		// Séparer le nom de la ville et du pays
		parts := strings.Split(location, "-")
		ville := parts[0]
		pays := parts[1]
		// Transformer "_"" en " " dans les noms de ville et de pays
		ville = strings.ReplaceAll(ville, "_", " ")
		pays = strings.ReplaceAll(pays, "_", " ")
		// Vérifier si la ville et le pays ne sont pas dans les suggestions
		if !containschaine(suggestions, ville) && !containschaine(suggestions, pays) {
			return ville, pays
		} else if !containschaine(suggestions, ville) && containschaine(suggestions, pays) {
			return ville, ""
		} else if !containschaine(suggestions, pays) && containschaine(suggestions, ville) {
			return pays, ""
		}
	}
	return "", "" // Retourner des valeurs vides si aucune suggestion n'est trouvée
}

func containschaine(tabstring []string, s string) bool {
	for _, c := range tabstring {
		if c == s {
			return true
		}
	}
	return false
}

func filterDataByRelations(relationsData []struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}, date, location, search string) ([]ArtistsInfo, error) {
	var filteredArtists []ArtistsInfo
	id, _ := recupIdByArtist(search)
	// Filtrer les données en fonction des combinaisons de filtres
	for _, artist := range relationsData {
		if id > 0 && date != "" && location == "" {
			if artist.ID == id {
				for _, dates := range artist.DatesLocations {
					for _, artistdate := range dates {
						if date == artistdate {
							artistInfo, err := recupArtistesByID(artist.ID)
							log.Println(artistInfo)
							if err != nil {
								return nil, err
							}
							filteredArtists = append(filteredArtists, artistInfo)
							break
						}
					}
				}
			}
		} else if id > 0 && date == "" && location != "" {
			if artist.ID == id {
				for locations := range artist.DatesLocations {
					if strings.Contains(locations, location) {
						artistInfo, err := recupArtistesByID(artist.ID)
						log.Println(artistInfo)
						if err != nil {
							return nil, err
						}
						filteredArtists = append(filteredArtists, artistInfo)
						break
					}
				}
			}
		} else {
			for loc, dates := range artist.DatesLocations {
				if strings.Contains(loc, location) {
					for _, artistdates := range dates {
						if date == artistdates {
							artistInfo, err := recupArtistesByID(artist.ID)
							log.Println(artistInfo)
							if err != nil {
								return nil, err
							}
							filteredArtists = append(filteredArtists, artistInfo)
							break
						}
					}
				}
			}
		}
	}
	return filteredArtists, nil
}

func filterDataBySearch(data []ArtistsInfo, search string) ([]ArtistsInfo, error) {
	if search == "" {
		return data, nil
	}
	var filteredData []ArtistsInfo
	for _, artist := range data {
		if strings.Contains(strings.ToLower(artist.Name), strings.ToLower(search)) {
			filteredData = append(filteredData, artist)
		} else if len(artist.Members) == 1 {
			if strings.Contains(strings.ToLower(artist.Members[0]), strings.ToLower(search)) {
				filteredData = append(filteredData, artist)
			}
		} else {
			for _, membre := range artist.Members {
				if strings.Contains(strings.ToLower(membre), strings.ToLower(search)) {
					filteredData = append(filteredData, artist)
				}
			}
		}
	}
	return filteredData, nil
}

func filterDataByYear(data []ArtistsInfo, years int) ([]ArtistsInfo, error) {
	var filterData []ArtistsInfo
	for _, artist := range data {
		if artist.CreationDate == years {
			filterData = append(filterData, artist)
		}

	}
	return filterData, nil
}

func filterDatabyMembers(data []ArtistsInfo, members int) ([]ArtistsInfo, error) {
	var count int
	var filterData []ArtistsInfo
	for _, artist := range data {
		count = len(artist.Members)
		if count == members {
			filterData = append(filterData, artist)
		}
	}
	return filterData, nil
}

func filterDatabyFirstAlbum(data []ArtistsInfo, dates string) ([]ArtistsInfo, error) {
	if dates == "" {
		return data, nil
	}
	var Results []ArtistsInfo
	for _, artist := range data {
		if dates == artist.FirstAlbum {
			Results = append(Results, artist)
			log.Print(Results, " premier album : ", dates)
		}
	}
	return Results, nil
}

func filterDataByDate(data []struct {
	ID    int      "json:\"id\""
	Dates []string "json:\"dates\""
}, filtre string) ([]ArtistsInfo, error) {
	if filtre == "" {
		return nil, nil
	}
	var artistsPlaying []ArtistsInfo
	filtreNew := strings.ToLower(strings.TrimLeft(filtre, "*")) // Enlevez le "*" et convertissez en minuscules
	for _, index := range data {
		for _, date := range index.Dates {
			datename := strings.ToLower(strings.TrimLeft(date, "*")) // Enlevez le "*" et convertissez en minuscules
			if datename == filtreNew {
				// Si la date correspond, récupérez les informations sur l'artiste à partir de l'ID de l'index
				artistInfo, err := recupArtistesByID(index.ID)
				log.Println(artistInfo)
				if err != nil {
					return nil, err
				}
				// Ajoutez les informations de l'artiste à la liste des artistes jouant à cette date
				artistsPlaying = append(artistsPlaying, artistInfo)
			}
		}
	}
	return artistsPlaying, nil
}

func filterDataByLocations(data []struct {
	ID        int      "json:\"id\""
	Locations []string "json:\"locations\""
	Dates     string   "json:\"dates\""
}, filtre string) ([]ArtistsInfo, error) {
	if filtre == "" {
		return nil, nil
	}
	var artistsPlaying []ArtistsInfo
	// Utilisez une carte pour suivre les artistes déjà trouvés
	artistsMap := make(map[int]bool)
	filtreNew := strings.ToLower(filtre)
	// Divisez la chaîne de recherche en mots individuels
	searchWords := strings.Fields(filtreNew)
	for _, index := range data {
		for _, location := range index.Locations {
			// Convertir la location en minuscules pour une correspondance insensible à la casse
			location = strings.ToLower(location)
			// Divisez le nom de lieu en mots individuels
			locationWords := strings.FieldsFunc(location, func(r rune) bool {
				return r == '_' || r == '-'
			})
			// Vérifiez si au moins un des mots de la recherche est présent dans le nom de lieu
			for _, word := range searchWords {
				for _, locWord := range locationWords {
					if strings.Contains(locWord, word) {
						// Vérifiez si l'artiste correspondant n'a pas déjà été ajouté à la carte
						if _, ok := artistsMap[index.ID]; !ok {
							// Si une correspondance partielle est trouvée, récupérez les informations sur l'artiste à partir de l'ID de l'index
							artistInfo, err := recupArtistesByID(index.ID)
							if err != nil {
								return nil, err
							}
							// Ajoutez les informations de l'artiste à la liste des artistes jouant à cette location
							artistsPlaying = append(artistsPlaying, artistInfo)
							// Ajoutez l'ID de l'artiste à la carte
							artistsMap[index.ID] = true
						}
						break
					}
				}
			}
		}
	}
	return artistsPlaying, nil
}

func recupArtistesByID(id int) (ArtistsInfo, error) {
	apiInfo, err := recupJSON()
	if err != nil {
		log.Printf("Erreur de récupération des infos API : %v\n", err)
		return ArtistsInfo{}, err
	}
	artistList, err := recupArtistes(apiInfo.ArtistsInfo)
	if err != nil {
		log.Printf("Erreur lors de la récupération des artistes : %v\n", err)
		return ArtistsInfo{}, err
	}
	for _, artist := range artistList {
		if artist.ID == id {
			return artist, nil
		}
	}
	return ArtistsInfo{}, fmt.Errorf("Aucun artiste trouvé avec l'ID %d", id)
}

func recupIdByArtist(name_artist string) (int, error) {
	apiInfo, err := recupJSON()
	if err != nil {
		log.Printf("Erreur de récupération des infos Api : %v\n", err)
		return 0, err
	}
	artistList, err := recupArtistes(apiInfo.ArtistsInfo)
	if err != nil {
		log.Printf("Erreur lors de la récupération des artistes : %v\n", err)
		return 0, err
	}
	for _, artist := range artistList {
		if artist.Name == name_artist {
			return artist.ID, nil
		}
	}
	return 0, fmt.Errorf("Aucun artiste trouvé avec le nom %s", name_artist)
}

func recupArtistes(url string) ([]ArtistsInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Erreur lors de la requête GET : %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	var artistsInfo []ArtistsInfo
	err = json.NewDecoder(resp.Body).Decode(&artistsInfo)
	if err != nil {
		log.Printf("Erreur lors du décodage JSON : %v\n", err)
		return nil, err
	}
	return artistsInfo, nil
}

func recupDates(url string) (*DatesInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Erreur lors de la requête GET : %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	var datesInfo DatesInfo
	err = json.NewDecoder(resp.Body).Decode(&datesInfo)
	if err != nil {
		log.Printf("Erreur lors du décodage JSON : %v\n", err)
		return nil, err
	}
	return &datesInfo, nil
}

func recupRelation(url string) (*RelationsInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Erreur lors de la requête GET : %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	var relationsInfo RelationsInfo
	err = json.NewDecoder(resp.Body).Decode(&relationsInfo)
	if err != nil {
		log.Printf("Erreur lors du décodage JSON : %v\n", err)
		return nil, err
	}
	return &relationsInfo, nil
}

func recupLocation(url string) (*LocationsInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Erreur lors de la requête GET : %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	var locationsInfo LocationsInfo
	err = json.NewDecoder(resp.Body).Decode(&locationsInfo)
	if err != nil {
		log.Printf("Erreur lors du décodage JSON : %v\n", err)
		return nil, err
	}
	return &locationsInfo, nil
}

func artist(api string) {
	artistList, err := recupArtistes(api)
	if err != nil {
		log.Printf("Erreur : %v\n", err)
		return
	}

	var saisie string
	fmt.Print("Quel artiste voulez-vous consulter ? (Saisissez le nom de l'artiste) (all): ")
	_, err = fmt.Scan(&saisie)
	if err != nil {
		log.Printf("Erreur de saisie : %v\n", err)
		return
	}

	var findArtiste *ArtistsInfo
	for _, artiste := range artistList {
		if strings.Contains(strings.ToLower(artiste.Name), strings.ToLower(saisie)) {
			findArtiste = &artiste
			break
		}
	}

	if findArtiste != nil {
		fmt.Printf("Info sur l'artiste %s:\nID: %d\n", saisie, findArtiste.ID)
		var saisieUser string
		fmt.Print("Que voulez-vous consulter ? (members) (creationDate) (firstAlbum) :")
		_, err = fmt.Scan(&saisieUser)
		if err != nil {
			log.Printf("Erreur de saisie : %v\n", err)
		}

		grp := len(findArtiste.Members) - 1
		if saisieUser == "members" {
			if grp == 1 {
				fmt.Printf("L'artiste est %s", findArtiste.Members[0])
			} else {
				fmt.Println("Les membres sont :")
				for _, membres := range findArtiste.Members {
					fmt.Println(membres)
				}
			}
		} else if saisieUser == "creationDate" {
			if grp == 1 {
				fmt.Printf("Le début de l'artiste est %d", findArtiste.CreationDate)
			} else {
				fmt.Printf("Le début du groupe est %d", findArtiste.CreationDate)
			}
		} else if saisieUser == "firstAlbum" {
			if grp == 1 {
				fmt.Printf("Le premier Album de l'artiste est sortie le %s", findArtiste.FirstAlbum)
			} else {
				fmt.Printf("Le premier Album du groupe est sortie le %s", findArtiste.FirstAlbum)
			}
		} else {
			artist(api)
		}
	} else {
		artist(api)
	}
}

func artistefindwithdate(api string, apiartist string) {
	var indexartist []int
	datesInfo, err := recupDates(api)
	if err != nil {
		log.Printf("Erreur : %v\n", err)
		return
	}

	var saisie string
	fmt.Print("Veuillez entrer une date (format : JJ-MM-AA ): ")
	fmt.Scan(&saisie)

	for i, index := range datesInfo.Index {
		for _, date := range index.Dates {
			if strings.HasPrefix(date, "*") {
				date = strings.TrimPrefix(date, "*")
				if date == saisie {
					indexartist = append(indexartist, i)
				}
			} else {
				if date == saisie {
					indexartist = append(indexartist, i)
				}
			}
		}
	}
	artistId(apiartist, indexartist)
}

func artistId(apiartist string, index []int) {
	artistList, err := recupArtistes(apiartist)
	if err != nil {
		log.Printf("Erreur : %v\n", err)
		return
	}

	for _, i := range index {
		artist := artistList[i]
		fmt.Println("Voici les informations sur l'artiste :", artist.Name)
		fmt.Println("ID :", artist.ID)
		fmt.Println("Image :", artist.Image)
		fmt.Println("Name :", artist.Name)
	}
}

func all(apiInfo *GroupieTracker) {
	artistList, err := recupArtistes(apiInfo.ArtistsInfo)
	if err != nil {
		log.Printf("Erreur : %v\n", err)
		return
	}

	locationsList, err := recupLocation(apiInfo.LocationsInfo)
	if err != nil {
		log.Printf("Erreur : %v\n", err)
		return
	}

	concertDatesList, err := recupDates(apiInfo.DatesInfo)
	if err != nil {
		log.Printf("Erreur : %v\n", err)
		return
	}

	relationList, err := recupRelation(apiInfo.RelationsInfo)
	if err != nil {
		log.Printf("Erreur : %v\n", err)
		return
	}

	for _, artiste := range artistList {
		fmt.Printf("Id: %d\n", artiste.ID)
		fmt.Printf("Image: %s\n", artiste.Image)
		fmt.Printf("Name: %s\n", artiste.Name)
		fmt.Printf("Members: %s\n", artiste.Members)
		fmt.Printf("CreationDate: %d\n", artiste.CreationDate)
		fmt.Printf("FirstAlbum: %s\n", artiste.FirstAlbum)

		// Vérifier la validité de l'index pour locationsList.Index
		if artiste.ID < len(locationsList.Index) {
			fmt.Printf("Locations: %s\n", locationsList.Index[artiste.ID].Locations)
		} else {
			fmt.Println("Locations: N/A")
		}

		// Vérifier la validité de l'index pour concertDatesList.Index
		if artiste.ID < len(concertDatesList.Index) {
			fmt.Printf("ConcertDates: %s\n", concertDatesList.Index[artiste.ID].Dates)
		} else {
			fmt.Println("ConcertDates: N/A")
		}

		// Vérifier la validité de l'index pour relationList.Index
		if artiste.ID < len(relationList.Index) {
			fmt.Printf("Relations: %s\n", relationList.Index[artiste.ID].DatesLocations)
		} else {
			fmt.Println("Relations: N/A")
		}

		fmt.Println("")
	}
}

func start() {
	apiInfo, err := recupJSON()
	if err != nil {
		log.Printf("Problème de récupération des infos API : %v\n", err)
		return
	}

	var saisie string
	fmt.Print("Que voulez-vous consulter ? (Artists) (Dates) (Locations) (Relations) (All): ")
	_, err = fmt.Scan(&saisie)
	if err != nil {
		log.Printf("Erreur de saisie : %v\n", err)
		return
	}

	switch saisie {
	case "Artists":
		fmt.Println("URL des artistes:", apiInfo.ArtistsInfo)
		artist(apiInfo.ArtistsInfo)
	case "Dates":
		fmt.Println("URL des artistes:", apiInfo.DatesInfo)
		artistefindwithdate(apiInfo.DatesInfo, apiInfo.ArtistsInfo)
	case "Locations":
		fmt.Println("URL des locations:", apiInfo.LocationsInfo)
	case "Relations":
		fmt.Println("URL des relations:", apiInfo.RelationsInfo)
	case "All":
		all(apiInfo)
	default:
		fmt.Println("Entrée non reconnue")
		start()
	}
}

func trier_ordre_alphabe(api []ArtistsInfo) ([]ArtistsInfo, error) {
	filterData := make([]ArtistsInfo, len(api))
	copy(filterData, api)

	sort.Slice(filterData, func(i, j int) bool {
		return filterData[i].Name < filterData[j].Name
	})

	if len(filterData) == 0 {
		log.Printf("La liste est vide")
	}
	return filterData, nil
}

// Assurez-vous d'avoir importé "sort" et "time"

func trier_ordre_concert_récent(data []struct {
	ID    int      "json:\"id\""
	Dates []string "json:\"dates\""
}) ([]ArtistsInfo, error) {
	date := make(map[string]int) // Carte stocke dates et id correspondants
	var min string
	for _, index := range data {
		if len(index.Dates) > 0 {
			min = index.Dates[0]
		}
		for _, dates := range index.Dates {
			if dates > min {
				min = dates
			}
		}
		date[min] = index.ID
	}

	var triagedates []string
	for dates := range date {
		triagedates = append(triagedates, dates)
	}
	sort.Slice(triagedates, func(i, j int) bool {
		return triagedates[i] > triagedates[j]
	})

	var Results []ArtistsInfo
	for _, dates := range triagedates {
		artistID := date[dates]
		artistInfo, _ := recupArtistesByID(artistID)
		Results = append(Results, artistInfo)
	}

	return Results, nil
}

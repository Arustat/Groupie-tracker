package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"html/template"
)

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
}

type LocationsInfo struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
		Dates     string   `json:"dates"`
	} `json:"index"`
}

type DatesInfo struct {
	Index []struct {
		ID    int      `json:"id"`
		Dates []string `json:"dates"`
	} `json:"index"`
}

type RelationsInfo struct {
	Index []struct {
		ID             int                  `json:"id"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}

type GroupieTracker struct {
	ArtistsInfo    string `json:"artists"`
	LocationsInfo  string `json:"locations"`
	DatesInfo      string `json:"dates"`
	RelationsInfo  string `json:"relation"`
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

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer la valeur du paramètre "search" dans l'URL
	search := r.URL.Query().Get("search")

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

	filterdataa := filterDataBySearch(artistList, search)

	// Créer un nouveau template à partir du fichier HTML
	tmpl, err := template.ParseFiles(htmlTemplatePath)
	if err != nil {
		http.Error(w, "Erreur lors de la création du template", http.StatusInternalServerError)
		return
	}

	// Exécuter le template en passant la liste des artistes filtrés
	err = tmpl.Execute(w, filterdataa)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution du template", http.StatusInternalServerError)
		return
	}
}


func filterDataBySearch(data []ArtistsInfo, search string) []ArtistsInfo {
	if search == "" {
		return data
	}
	var filteredData []ArtistsInfo
	for _, artist := range data {
		artistName := strings.ToLower(artist.Name)
		searchTerm := strings.ToLower(search)
		if strings.Contains(artistName, searchTerm) {
			filteredData = append(filteredData, artist)
		}
		fmt.Println("Résultats de la recherche :", filteredData)
	}
	return filteredData
}

func main() {
	start()
	
	// Intégre le dossier asset dans le serveur
	http.Handle("/asset/", http.StripPrefix("/asset/", http.FileServer(http.Dir("asset"))))
	
	// Définit la route principale
	http.HandleFunc("/", indexHandler)
	
	// Définit la route de recherche
	http.HandleFunc("/search", searchHandler)
	
	// Lance le serveur
	log.Fatal(http.ListenAndServe(":8010", nil))
	
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


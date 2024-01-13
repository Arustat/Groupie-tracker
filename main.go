package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ArtistsInfo struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"CreationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

type LocationsInfo struct {
    ID        int      `json:"id"`
    Locations []string `json:"locations"`
    Dates     []string `json:"dates"`
}

type DatesInfo struct {
    ID    int      `json:"id"`
    Dates []string `json:"dates"`
}

type RelationsInfo struct {
    ID             int         `json:"id"`
    DatesLocations []string    `json:"datesLocations"`
}

type GroupieTracker struct {
    ArtistsInfo string     `json:"artists"`
    Locations   string   `json:"locations"`
    Dates       string      `json:"dates"`
    Relations   string   `json:"relation"`
}



func recupJSON() (*GroupieTracker, error) {
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api")
	if err != nil {
		fmt.Println("Erreur lors de la requête GET :", err)
		return nil, err
	}
	defer resp.Body.Close()

	var apiInfo GroupieTracker
	err = json.NewDecoder(resp.Body).Decode(&apiInfo)
	if err != nil {
		return nil, err
	}
	return &apiInfo, nil
}

func main() {
	start()
}

func start() {
	apiInfo, err := recupJSON()
	if err != nil {
		fmt.Println("Problème de récupération des infos API", err)
		return
	}
	var saisie string
	fmt.Print("Que voulez-vous consulter ? (Artists) (Dates) (Locations) (Relations): ")
	_, err = fmt.Scan(&saisie)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}

	switch saisie {
	case "Artists":
		fmt.Println("URL des artistes:", apiInfo.ArtistsInfo)
		artist(apiInfo.ArtistsInfo)
	case "Dates":
		fmt.Println("URL des dates:", apiInfo.Dates)
		datesList, err := recupDates(apiInfo.Dates)
		if err != nil {
			fmt.Println("Erreur",err)
			return
		}
		var dateSaisie string
		fmt.Print("Quelle date voulez-vous consulter ? (Saisissez la Date) :")
		_, err = fmt.Scan(&dateSaisie)
        if err != nil {
            fmt.Println("Erreur", err)
            return
        }

        searchByDate(datesList, dateSaisie)
	case "Locations":
		fmt.Println("URL des locations:", apiInfo.Locations)
	case "Relations":
		fmt.Println("URL des relations:", apiInfo.Relations)
	default:
		fmt.Println("Entrée non reconnue")
		start()
	}
}

func recupArtistes(url string) ([]ArtistsInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artistsInfo []ArtistsInfo
	err = json.NewDecoder(resp.Body).Decode(&artistsInfo)
	if err != nil {
		return nil, err
	}
	return artistsInfo, nil
}

func artist(api string) {
	artistList, err := recupArtistes(api)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}

	var saisie string
	fmt.Print("Quel artiste voulez-vous consulter ? (Saisissez le nom de l'artiste) : ")
	_, err = fmt.Scan(&saisie)
	if err != nil {
		fmt.Println("Erreur", err)
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
        var saisie_user string
        fmt.Print("Que voulez vous consulter ? (members) (creationDate) (firstAlbum) :")
        _, err = fmt.Scan(&saisie_user)
	    if err != nil {
		    fmt.Println("Erreur", err)
	    }
        var grp int
        grp = len(findArtiste.Members) - 1
        if saisie_user == "members"{
            if grp == 1{
                fmt.Println("L'artiste est %s", findArtiste.Members[0])
            }else{
                fmt.Println("Les membres sont :")
                for _, membres := range findArtiste.Members{
                    fmt.Println(membres)
                }
            }
        }else if saisie_user =="creationDate"{
            if grp == 1 {
                fmt.Printf("Le début de l'artiste est %d",findArtiste.CreationDate)
            }else{
                fmt.Printf("Le début du groupe est %d",findArtiste.CreationDate)
            }
        }else if saisie_user =="firstAlbum"{
            if grp == 1 {
                fmt.Printf("Le premier Album de l'artiste est sortie le %s",findArtiste.FirstAlbum)
            }else{
                fmt.Printf("Le premier Album du groupe est sortie le %s",findArtiste.FirstAlbum)
            }
        }else{
            artist(api)
        }
	} else {
		artist(api)
	}
}

func recupDates(url string)([]DatesInfo, error){
	resp, err := http.Get(url)
	if err != nil {
		return nil,err
	}
	defer resp.Body.Close()

	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("La requête a retourné un statut non OK: %d", resp.StatusCode)
	}

	var datesInfo []DatesInfo
	err = json.NewDecoder(resp.Body).Decode(&datesInfo)
	if err != nil{
		return nil,err
	}
	return datesInfo, nil
}

func searchByDate(datesList []DatesInfo, date string){
	found := false

	for _, d := range datesList{
		for _,dateItem := range d.Dates{
			found = true
			getArtistByDate(d.ID,dateItem)
			break
		}
		if found {
			break
		}
	}
	if !found{
		fmt.Printf("Aucun artiste performe pour la date %s \n",date)
	}
}

func getArtistByDate(dateID int, date string){
	fmt.Printf("Artistes qui performent à la date %s (ID: $d):\n",date,dateID)
}
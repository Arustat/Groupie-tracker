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
	Index []struct  {
    ID        int      `json:"id"`
    Locations []string `json:"locations"`
    Dates     string `json:"dates"`
	}`json: "index"`
}

type DatesInfo struct {
	Index []struct  {
		Id 		int    `json: "id"`
		Dates []string `json: "dates"`
	}`json: "index"`
}

type RelationsInfo struct {
	Index []struct {
		ID             int         `json:"id"`
    	DatesLocations map[string][]string    `json:"datesLocations"`
	}`json: index`
}

type GroupieTracker struct {
    ArtistsInfo string     `json:"artists"`
    LocationsInfo   string   `json:"locations"`
    DatesInfo   string `json:"dates"`
    RelationsInfo   string   `json:"relation"`
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
	fmt.Print("Que voulez-vous consulter ? (Artists) (Dates) (Locations) (Relations) (All): ")
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
		fmt.Println("URL des artistes:", apiInfo.DatesInfo)
		artistefindwithdate(apiInfo.DatesInfo, apiInfo.ArtistsInfo)
	case "Locations":
		fmt.Println("URL des locations:", apiInfo.LocationsInfo)
	case "Relations":
		fmt.Println("URL des relations:", apiInfo.RelationsInfo)
	case "All":
		all()
	default:
		fmt.Println("Entrée non reconnue")
		start()
	}
}

func all(){
	api, err := recupJSON()
	if err != nil {
		fmt.Println("Problème de récupération des infos API", err)
		return
	}
	artistList, err := recupArtistes(api.ArtistsInfo)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}
	locationsList, err := recupLocation(api.LocationsInfo)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}
	concertDatesList, err := recupDates(api.DatesInfo)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}

	relationList, err := recupRelation(api.RelationsInfo)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}

	for _, artiste := range artistList {
		fmt.Printf("Id: %d",artiste.ID)
		fmt.Println("")
		fmt.Printf("Image: %s",artiste.Image)
		fmt.Println("")
		fmt.Printf("Name: %s",artiste.Name)
		fmt.Println("")
		fmt.Printf("Members: %s",artiste.Members)
		fmt.Println("")
		fmt.Printf("CreationDate: %d",artiste.CreationDate)
		fmt.Println("")
		fmt.Printf("FirstAlbum: %s",artiste.FirstAlbum)
		fmt.Println("")
		fmt.Printf("Locations: %s",locationsList.Index[artiste.ID].Locations)
		fmt.Println("")
		fmt.Printf("ConcertDates: %s",concertDatesList.Index[artiste.ID].Dates)
		fmt.Println("")
		fmt.Printf("Relations: %s",relationList.Index[artiste.ID].DatesLocations)
		fmt.Println("") 
		fmt.Println("")
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

func recupDates(url string) (*DatesInfo, error){
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var datesInfo DatesInfo
	err = json.NewDecoder(resp.Body).Decode(&datesInfo)
	if err != nil {
		return nil, err
	}
	return &datesInfo, nil
}

func recupRelation(url string) (*RelationsInfo, error){
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var relationsInfo RelationsInfo
	err = json.NewDecoder(resp.Body).Decode(&relationsInfo)
	if err != nil {
		return nil, err
	}
	return &relationsInfo, nil
}

func recupLocation(url string) (*LocationsInfo, error){
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var locationsInfo LocationsInfo
	err = json.NewDecoder(resp.Body).Decode(&locationsInfo)
	if err != nil {
		return nil, err
	}
	return &locationsInfo, nil
}


func artist(api string) {
	artistList, err := recupArtistes(api)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}

	var saisie string
	fmt.Print("Quel artiste voulez-vous consulter ? (Saisissez le nom de l'artiste) (all): ")
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

func artistefindwithdate(api string, apiartist string){
	var indexartist []int
	datesInfo, err := recupDates(api)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}

	var saisie string
	fmt.Print("Veuillez entrer une date (format : JJ-MM-AA ): ")
	fmt.Scan(&saisie)

	for i, index := range datesInfo.Index {
		for _, date := range index.Dates {
			if string(date[0]) == "*"{
				date = strings.Replace(date, string("*"),"",-1)
				if date == saisie{
					indexartist = append(indexartist, i)
				}
			}else{
				if date == saisie{
					indexartist = append(indexartist, i)
				}
			}
		}
	}
	artistId(apiartist, indexartist)
}

func artistId(apiartist string, index []int){
	artistList, err := recupArtistes(apiartist)
	if err != nil {
		fmt.Println("Erreur", err)
		return
	}

	for _,i := range index{
		artist := artistList[i]
		fmt.Println("Voici les informations sur l'artiste :", artist.Name)
		fmt.Println("ID :", artist.ID)
		fmt.Println("Image :", artist.Image)
		fmt.Println("Name :", artist.Name)
	}
}
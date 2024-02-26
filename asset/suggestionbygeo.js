$(document).ready(function() {
    $("#autocompletegeo").on("input", function() {
        var query = $(this).val();
        $.ajax({
            url: "/suggestgeo?query=" + query,
            method: "GET",
            success: function(data) {
                $("#suggestionsloc").empty();
                var suggestionA = data.split(",");
                console.log("Suggestions array:", suggestionA);
                if (suggestionA.length === 1 && suggestionA[0] === "null\n") {
                    $("#suggestionsloc").append("<li>No artist matching!</li>");
                } else {
                    // Ajouter l'élément "Géolocalisation" en premier
                    $("#suggestionsloc").append("<li id='geocalisationItem'>Géolocalisation</li>");
                    suggestionA.forEach(function(suggestion) {
                        suggestion = suggestion.replace(/["\[\]]/g, '');
                        $("#suggestionsloc").append("<li>" + suggestion + "</li>");
                    });
                }
            },
            error: function() {
                // Gérer les erreurs
                $("#suggestionsloc").empty().append("<li>Error fetching suggestions</li>");
            }
        });
    });
    
    $(document).on("click", "#suggestionsloc li", function() {
        var suggestion = $(this).text();
        if (suggestion === "Géolocalisation") {
            geo(); // Appeler la fonction geo() lorsque l'utilisateur sélectionne "Géolocalisation"
        } else {
            $("#autocompletegeo").val(suggestion);
            $("#suggestionsloc").empty(); // Vide les suggestions après que l'utilisateur en a sélectionné une
        }
    });
    // Gestionnaire d'événements pour fermer les suggestions en dehors de la zone de suggestion
    $(document).on("click", function(event) {
        if (!$(event.target).closest("#suggestionsloc").length && !$(event.target).is("#autocompletegeo")) {
            $("#suggestionsloc").empty();
        }
    });
});

function geo() {
    if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(function(position) {
            var latitudeuser = position.coords.latitude;
            var longitudeuser = position.coords.longitude;
            console.log('Latitude : ', latitudeuser);
            console.log('Longitude : ', longitudeuser);
            
            obtenirNomVillePays(latitudeuser, longitudeuser, function(city, country) {
                $("#suggestionsloc").empty(); // Vider les suggestions
                // Ajouter l'élément "Géolocalisation" en premier
                $("#suggestionsloc").append("<li>"+ city +"</li>");
                $("#suggestionsloc").append("<li>"+ country +"</li>");
            });
        }, function(error) {
            console.error('Erreur lors de la récupération de la position : ', error.message);
        });
    } else {
        console.error("La géolocalisation n'est pas prise en charge par ce navigateur.");
    }

    function obtenirNomVillePays(latitude, longitude, callback) {
        var url = `/geonames?lat=${latitude}&lng=${longitude}`;
        
        fetch(url)
            .then(response => response.json())
            .then(data => {
                if (data.geonames.length > 0) {
                    var city = data.geonames[0].name;
                    var country = data.geonames[0].countryName;
                    console.log('Ville : ', city);
                    console.log('Pays : ', country);
                    callback(city, country);
                } else {
                    console.log('Aucune information trouvée.');
                }
            })
            .catch(error => console.error('Erreur lors de la récupération du nom de la ville/pays : ', error));
    }
}

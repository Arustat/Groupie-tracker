$(document).ready(function() {
    $("#autocomplete").on("input", function() {
        var query = $(this).val();
        $.ajax({
            url: "/suggest?query=" + query,
            method: "GET",
            success: function(data) {
                $("#suggestions").empty();
                var suggestionA = data.split(",");
                console.log("Suggestions array:", suggestionA);
                if (suggestionA.length === 1 && suggestionA[0] === "null\n") {
                    $("#suggestions").append("<li>No artist matching!</li>");
                } else {
                    suggestionA.forEach(function(suggestion) {
                        suggestion = suggestion.replace(/["\[\]]/g, '');
                        $("#suggestions").append("<li>" + suggestion + "</li>");
                    });
                }
                
            },
            error: function() {
                // Gérer les erreurs
                $("#suggestions").empty().append("<li>Error fetching suggestions</li>");
            }
        });
    });

    $(document).on("click", "#suggestions li", function() {
        var suggestion = $(this).text();
        $("#autocomplete").val(suggestion);
        $("#suggestions").empty(); // Vide les suggestions après que l'utilisateur en a sélectionné une
    });
    // Gestionnaire d'événements pour fermer les suggestions en dehors de la zone de suggestion
    $(document).on("click", function(event) {
        if (!$(event.target).closest("#suggestions").length && !$(event.target).is("#autocomplete")) {
            $("#suggestions").empty();
        }
    });
});
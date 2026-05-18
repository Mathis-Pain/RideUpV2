// Script spécifique pour la page de création d'événement moto
document.addEventListener("DOMContentLoaded", function () {
  const mapDiv = document.getElementById("map");

  // Récupère les coordonnées du dataset HTML
  const lat = parseFloat(mapDiv.dataset.lat) || 48.8566;
  const lon = parseFloat(mapDiv.dataset.lon) || 2.3522;

  // Initialisation de la carte Leaflet
  const map = L.map("map").setView([lat, lon], 13);

  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 19,
    attribution: "&copy; OpenStreetMap contributors",
  }).addTo(map);

  // Marqueur de position actuelle (bleu)
  L.marker([lat, lon]).addTo(map).bindPopup("📍 Votre position").openPopup();

  // Variable pour stocker le marqueur de départ
  let startMarker = null;

  // Fonction pour placer le marqueur et mettre à jour les champs
  function setStartMarker(latitude, longitude, label = "Point de départ") {
    // Supprimer le marqueur précédent s'il existe
    if (startMarker) {
      map.removeLayer(startMarker);
    }

    // Ajouter un nouveau marqueur rouge
    startMarker = L.marker([latitude, longitude], {
      icon: L.icon({
        iconUrl:
          "https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-red.png",
        shadowUrl:
          "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png",
        iconSize: [25, 41],
        iconAnchor: [12, 41],
        popupAnchor: [1, -34],
        shadowSize: [41, 41],
      }),
    })
      .addTo(map)
      .bindPopup(`🏍️ ${label}`)
      .openPopup();

    // CRUCIAL : Mettre à jour les champs cachés pour le backend
    document.getElementById("latitude").value = latitude.toFixed(6);
    document.getElementById("longitude").value = longitude.toFixed(6);

    console.log(
      `✅ Coordonnées définies: ${latitude.toFixed(6)}, ${longitude.toFixed(6)}`
    );
  }

  // Événement : double-clic sur la carte pour placer le point de départ
  map.on("dblclick", function (e) {
    const latitude = e.latlng.lat;
    const longitude = e.latlng.lng;

    // Placer le marqueur temporairement
    setStartMarker(latitude, longitude, "Recherche de l'adresse...");

    // Géocodage inversé pour obtenir l'adresse
    reverseGeocode(latitude, longitude);

    console.log("🗺️ Point placé via double-clic sur la carte");
  });

  // Géocodage de l'adresse saisie
  const addressInput = document.getElementById("address");
  let geocodeTimeout = null;

  addressInput.addEventListener("input", function () {
    clearTimeout(geocodeTimeout);

    const address = this.value.trim();

    // Attendre 1 seconde après la dernière frappe avant de géocoder
    if (address.length > 3) {
      geocodeTimeout = setTimeout(() => {
        geocodeAddress(address);
      }, 1000);
    }
  });

  // Fonction de géocodage via Nominatim OpenStreetMap
  async function geocodeAddress(address) {
    try {
      console.log(`🔍 Recherche de l'adresse: ${address}`);

      const response = await fetch(
        `https://nominatim.openstreetmap.org/search?` +
          `q=${encodeURIComponent(address)}` +
          `&format=json` +
          `&limit=5` + // Augmenté à 5 résultats
          `&addressdetails=1`, // Détails de l'adresse
        {
          headers: {
            "User-Agent": "RideUP/1.0 (moto-event-app)",
          },
        }
      );

      if (!response.ok) {
        throw new Error(`Erreur HTTP: ${response.status}`);
      }

      const data = await response.json();

      if (data && data.length > 0) {
        const result = data[0];
        const latitude = parseFloat(result.lat);
        const longitude = parseFloat(result.lon);

        // Construction de l'adresse complète avec numéro
        let fullAddress = result.display_name;
        if (result.address) {
          const addr = result.address;
          const parts = [];

          if (addr.house_number) parts.push(addr.house_number);
          if (addr.road) parts.push(addr.road);
          if (addr.village || addr.town || addr.city) {
            parts.push(addr.village || addr.town || addr.city);
          }
          if (addr.postcode) parts.push(addr.postcode);

          if (parts.length > 0) {
            fullAddress = parts.join(", ");
          }
        }

        console.log(`✅ Adresse trouvée: ${fullAddress}`);

        // Placer le marqueur avec l'adresse complète
        setStartMarker(latitude, longitude, fullAddress);

        // Mettre à jour le champ adresse avec l'adresse formatée
        document.getElementById("address").value = fullAddress;

        // Centrer la carte sur le résultat avec un zoom approprié
        map.setView([latitude, longitude], 16);
      } else {
        console.warn("❌ Aucun résultat trouvé pour cette adresse");
        alert(
          "⚠️ Adresse introuvable. Vérifiez l'orthographe ou ajoutez la ville.\nExemple: '10 rue de la Paix, Paris'"
        );
      }
    } catch (error) {
      console.error("❌ Erreur lors du géocodage:", error);
      alert("⚠️ Erreur de connexion. Vérifiez votre connexion internet.");
    }
  }

  // Fonction de géocodage inversé (coordonnées → adresse)
  async function reverseGeocode(latitude, longitude) {
    try {
      console.log(`🔍 Recherche de l'adresse pour: ${latitude}, ${longitude}`);

      const response = await fetch(
        `https://nominatim.openstreetmap.org/reverse?` +
          `lat=${latitude}` +
          `&lon=${longitude}` +
          `&format=json` +
          `&addressdetails=1` +
          `&zoom=18`, // Zoom max pour plus de précision
        {
          headers: {
            "User-Agent": "RideUP/1.0 (moto-event-app)",
          },
        }
      );

      if (!response.ok) {
        throw new Error(`Erreur HTTP: ${response.status}`);
      }

      const data = await response.json();

      if (data && data.address) {
        const addr = data.address;
        const parts = [];

        // Construction de l'adresse avec numéro
        if (addr.house_number) parts.push(addr.house_number);
        if (addr.road) parts.push(addr.road);
        if (addr.village || addr.town || addr.city) {
          parts.push(addr.village || addr.town || addr.city);
        }
        if (addr.postcode) parts.push(addr.postcode);

        const fullAddress =
          parts.length > 0 ? parts.join(", ") : data.display_name;

        console.log(`✅ Adresse trouvée: ${fullAddress}`);

        // Mettre à jour le marqueur avec l'adresse
        if (startMarker) {
          startMarker.setPopupContent(`🏍️ ${fullAddress}`);
          startMarker.openPopup();
        }

        // Mettre à jour le champ adresse
        document.getElementById("address").value = fullAddress;
      } else {
        console.warn(
          "❌ Impossible de trouver une adresse pour ces coordonnées"
        );
        const simpleAddr = `Lat: ${latitude.toFixed(
          6
        )}, Lon: ${longitude.toFixed(6)}`;
        if (startMarker) {
          startMarker.setPopupContent(`🏍️ ${simpleAddr}`);
        }
      }
    } catch (error) {
      console.error("❌ Erreur lors du géocodage inversé:", error);
    }
  }

  // Validation avant soumission du formulaire
  const form = document.querySelector(".form-newevent");

  if (form) {
    form.addEventListener("submit", function (e) {
      const lat = document.getElementById("latitude").value;
      const lon = document.getElementById("longitude").value;
      const address = document.getElementById("address").value.trim();

      // Vérifier qu'on a soit une adresse, soit des coordonnées
      if (!address && (!lat || !lon)) {
        e.preventDefault();
        alert(
          "⚠️ Veuillez saisir une adresse OU double-cliquer sur la carte pour définir le point de départ"
        );
        return false;
      }

      console.log("📤 Soumission du formulaire:");
      console.log("  - Adresse:", address || "(coordonnées directes)");
      console.log("  - Latitude:", lat || "(à géocoder)");
      console.log("  - Longitude:", lon || "(à géocoder)");
    });
  }
});

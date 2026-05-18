document.addEventListener("DOMContentLoaded", function () {
  const mapDiv = document.getElementById("map");

  // Position par défaut : si backend n'envoie rien, fallback sur Rouen
  let lat = parseFloat(mapDiv.dataset.lat);
  let lon = parseFloat(mapDiv.dataset.lon);
  if (isNaN(lat)) lat = 48.8566;
  if (isNaN(lon)) lon = 2.3522;

  const map = L.map("map").setView([lat, lon], 13);

  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 19,
    attribution: "&copy; OpenStreetMap contributors",
  }).addTo(map);

  // Marqueur de l'utilisateur
  const userMarker = L.marker([lat, lon], {draggable: true})
    .addTo(map)
    .bindPopup("📍 Votre position")
    .openPopup();

  // Cercle de rayon autour de l'utilisateur
  let radius = parseInt(document.getElementById("radius").value) * 1000;
  const circle = L.circle([lat, lon], {radius}).addTo(map);

  // Mettre à jour les champs cachés
  document.getElementById("latitude").value = lat.toFixed(6);
  document.getElementById("longitude").value = lon.toFixed(6);

  // Déplacer le marqueur met à jour le cercle et les champs
  userMarker.on("drag", function (e) {
    const pos = e.latlng;
    circle.setLatLng(pos);
    document.getElementById("latitude").value = pos.lat.toFixed(6);
    document.getElementById("longitude").value = pos.lng.toFixed(6);
  });

  // Double-clic sur la carte pour repositionner le marqueur et mettre à jour l'adresse
  map.on("dblclick", async function (e) {
    const pos = e.latlng;

    // Déplacer le marqueur et le cercle
    userMarker.setLatLng(pos);
    circle.setLatLng(pos);

    // Mettre à jour les champs cachés
    document.getElementById("latitude").value = pos.lat.toFixed(6);
    document.getElementById("longitude").value = pos.lng.toFixed(6);

    // Géocodage inversé pour récupérer l'adresse
    try {
      const response = await fetch(
        `https://nominatim.openstreetmap.org/reverse?lat=${pos.lat}&lon=${pos.lng}&format=json&addressdetails=1&zoom=18`,
        {headers: {"User-Agent": "RideUP/1.0 (moto-event-app)"}}
      );
      if (!response.ok) throw new Error("Erreur HTTP: " + response.status);

      const data = await response.json();
      let fullAddress = data.display_name || "";
      if (data.address) {
        const parts = [];
        if (data.address.house_number) parts.push(data.address.house_number);
        if (data.address.road) parts.push(data.address.road);
        if (data.address.village || data.address.town || data.address.city)
          parts.push(
            data.address.village || data.address.town || data.address.city
          );
        if (data.address.postcode) parts.push(data.address.postcode);
        if (parts.length > 0) fullAddress = parts.join(", ");
      }

      // Mettre à jour le champ adresse
      document.getElementById("address").value = fullAddress;
    } catch (err) {
      console.error("Erreur géocodage inversé:", err);
      document.getElementById("address").value = ""; // si erreur, vide le champ
    }
  });

  // Changement de rayon
  document.getElementById("radius").addEventListener("change", function () {
    const newRadius = parseInt(this.value) * 1000;
    circle.setRadius(newRadius);
  });
});

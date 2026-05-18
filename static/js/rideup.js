document.addEventListener("DOMContentLoaded", () => {
  const mapDiv = document.getElementById("map");
  const lat = parseFloat(mapDiv.dataset.lat) || 48.8566;
  const lon = parseFloat(mapDiv.dataset.lon) || 2.3522;
  const currentUserId = parseInt(mapDiv.dataset.userId);

  // Initialisation de la map
  const map = L.map("map").setView([lat, lon], 8);

  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 19,
    attribution: "&copy; OpenStreetMap contributors",
  }).addTo(map);

  L.marker([lat, lon]).addTo(map).bindPopup("📍 Votre position").openPopup();

  // Stockage des marqueurs
  const markers = {};

  try {
    const eventsData = mapDiv.dataset.events;
    console.log("🔍 Data brute:", eventsData);

    if (eventsData) {
      const events = JSON.parse(eventsData);
      console.log("✅ Events parsés:", events);
      console.log("📊 Nombre d'events:", events.length);

      events.forEach((event) => {
        console.log("📌 Traitement event:", event.id, event.title);

        if (event.latitude && event.longitude) {
          let iconColor;
          if (event.created_by === currentUserId) {
            iconColor = "green";
          } else if (event.user_joined) {
            iconColor = "violet";
          } else {
            iconColor = "red";
          }

          const iconUrl = `https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-${iconColor}.png`;

          const marker = L.marker([event.latitude, event.longitude], {
            icon: L.icon({
              iconUrl: iconUrl,
              iconSize: [25, 41],
              iconAnchor: [12, 41],
              popupAnchor: [1, -34],
              shadowUrl:
                "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png",
              shadowSize: [41, 41],
            }),
          }).addTo(map);

          // Stocke le marqueur
          markers[event.id] = marker;

          // 🔥 FORMATAGE DE LA DATE
          let formattedStart = "Non défini";
          if (event.start_datetime) {
            try {
              const date = new Date(event.start_datetime);
              if (!isNaN(date.getTime())) {
                formattedStart = date.toLocaleString("fr-FR", {
                  day: "2-digit",
                  month: "2-digit",
                  year: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                });
                console.log("✅ Date formatée:", formattedStart);
              } else {
                console.error("❌ Date invalide pour event", event.id);
              }
            } catch (error) {
              console.error("❌ Erreur formatage date:", error);
            }
          }

          // 🔥 CONSTRUCTION DE LA POPUP
          let popupContent = `
            <div style="max-width: 300px;">
              <h3 style="margin: 0 0 10px 0; color: #333;">${event.title}</h3>
              <p style="margin: 5px 0;"><strong>📅< Date :</strong> ${formattedStart}</p>
              <p style="margin: 5px 0;"><strong>📍 Lieu :</strong> ${
                event.address || "Non défini"
              }</p>
              <p style="margin: 5px 0;"><strong>👤 Organisateur :</strong> ${
                event.creator_name || "Inconnu"
              }</p>
          `;

          if (event.description) {
            popupContent += `<p style="margin: 10px 0 5px 0;"><strong>Description :</strong></p><p style="margin: 0;">${event.description}</p>`;
          }

          popupContent += `<p style="margin: 10px 0 5px 0;"><strong>👥 Participants :</strong> ${
            event.participants || 0
          }</p>`;

          // Boutons rejoindre/quitter/supprimer
          if (event.created_by === currentUserId) {
            popupContent += `
              <button class="delete-btn" data-event-id="${event.id}" style="
                margin-top: 10px;
                padding: 8px 16px;
                background-color: #dc3545;
                color: white;
                border: none;
                border-radius: 4px;
                cursor: pointer;
                font-size: 14px;
              ">🗑️ SUPPRIMER</button>
            `;
          } else {
            const buttonText = event.user_joined
              ? "❌ QUITTER"
              : "✅ REJOINDRE";
            const buttonColor = event.user_joined ? "#dc3545" : "#007bff";

            popupContent += `
              <button class="join-btn-popup ${
                event.user_joined ? "joined" : ""
              }" 
                      data-event-id="${event.id}" style="
                margin-top: 10px;
                padding: 8px 16px;
                background-color: ${buttonColor};
                color: white;
                border: none;
                border-radius: 4px;
                cursor: pointer;
                font-size: 14px;
              ">${buttonText}</button>
            `;
          }

          popupContent += `</div>`;

          marker.bindPopup(popupContent);
          console.log("✅ Marqueur créé pour event", event.id);
        } else {
          console.warn("⚠️ Event sans coordonnées:", event.id);
        }
      });

      console.log("✅ Total marqueurs créés:", Object.keys(markers).length);
    } else {
      console.warn("⚠️ Aucune donnée d'événements");
    }
  } catch (error) {
    console.error("❌ Erreur lors du parsing des events:", error);
  }

  // Événement clic sur une card (centrage sur le marker)
  document.querySelectorAll(".card").forEach((card) => {
    card.addEventListener("click", (e) => {
      if (
        e.target.classList.contains("join-btn") ||
        e.target.classList.contains("delete-btn") ||
        e.target.closest(".join-btn") ||
        e.target.closest(".delete-btn")
      ) {
        return;
      }

      const eventId = card.dataset.eventId;
      const cardLat = parseFloat(card.dataset.lat);
      const cardLon = parseFloat(card.dataset.lon);

      console.log("🎯 Clic sur card:", eventId, cardLat, cardLon);

      if (markers[eventId] && !isNaN(cardLat) && !isNaN(cardLon)) {
        map.setView([cardLat, cardLon], 16, {
          animate: true,
          duration: 0.5,
        });

        setTimeout(() => {
          markers[eventId].openPopup();
        }, 300);

        document.getElementById("map").scrollIntoView({
          behavior: "smooth",
          block: "center",
        });
      } else {
        console.error("❌ Marqueur introuvable:", eventId);
      }
    });
  });

  // Gestion des boutons
  document.addEventListener("click", async (e) => {
    if (
      e.target.classList.contains("join-btn") ||
      e.target.classList.contains("delete-btn") ||
      e.target.classList.contains("join-btn-popup")
    ) {
      const btn = e.target;
      const eventId = btn.dataset.eventId;
      const action = btn.textContent.trim().toLowerCase();

      console.log("🔘 Clic bouton:", action, "Event:", eventId);

      // Centrage sur la map avant l'action
      const card = btn.closest(".card");
      if (card) {
        const cardLat = parseFloat(card.dataset.lat);
        const cardLon = parseFloat(card.dataset.lon);

        if (markers[eventId] && !isNaN(cardLat) && !isNaN(cardLon)) {
          map.setView([cardLat, cardLon], 16, {
            animate: true,
            duration: 0.5,
          });

          setTimeout(() => {
            markers[eventId].openPopup();
          }, 300);

          document.getElementById("map").scrollIntoView({
            behavior: "smooth",
            block: "center",
          });
        }
      }

      try {
        if (action.includes("supprimer")) {
          if (!confirm("Voulez-vous vraiment supprimer cet événement ?"))
            return;

          const response = await fetch("/RideUp", {
            method: "POST",
            headers: {"Content-Type": "application/x-www-form-urlencoded"},
            body: `event_id=${eventId}&action=delete`,
          });

          if (!response.ok)
            throw new Error("Erreur serveur lors de la suppression");

          const data = await response.json();

          if (data.success) {
            window.location.href = "/RideUp";
          } else {
            alert("Impossible de supprimer cet événement.");
          }
          return;
        }

        const actionType = action.includes("rejoindre") ? "join" : "leave";
        const response = await fetch("/JoinEvent", {
          method: "POST",
          headers: {"Content-Type": "application/x-www-form-urlencoded"},
          body: `event_id=${eventId}&action=${actionType}`,
        });

        if (!response.ok) throw new Error("Erreur serveur");
        const data = await response.json();
        console.log("✅ Réponse serveur:", data);

        window.location.href = "/RideUp";
      } catch (err) {
        console.error("❌ Erreur:", err);
        alert("Une erreur est survenue, veuillez réessayer.");
      }
    }
  });
});

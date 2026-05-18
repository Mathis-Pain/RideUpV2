document.addEventListener("DOMContentLoaded", () => {
  const buttons = document.querySelectorAll(".join-btn");
  console.log("Nombre de boutons trouvés :", buttons.length);

  buttons.forEach((btn) => {
    btn.addEventListener("click", async () => {
      const eventId = btn.dataset.eventId;
      const action = btn.textContent.trim().toLowerCase(); // "rejoindre", "annuler" ou "supprimer"
      console.log("Action détectée :", action, "pour event", eventId);

      try {
        // 🔹 Suppression d’un événement (propriétaire)
        if (action === "supprimer") {
          if (!confirm("Voulez-vous vraiment supprimer cet événement ?"))
            return;

          const response = await fetch("/Admin", {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: `event_id=${eventId}&action=delete`,
          });

          if (!response.ok)
            throw new Error("Erreur serveur lors de la suppression");

          const data = await response.json();

          if (data.success) {
            // ✅ Option 1 : supprimer dynamiquement la carte
            const card = btn.closest(".card");
            if (card) card.remove();

            // ✅ Option 2 : recharger la page (décommenter si nécessaire)
            // window.location.href = "/RideUp";
          } else {
            alert("Impossible de supprimer cet événement.");
          }
        }
      } catch (err) {
        console.error("Erreur lors de la suppression :", err);
        alert("Une erreur est survenue. Vérifiez la console.");
      }
    });
  });
});

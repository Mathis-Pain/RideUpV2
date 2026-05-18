document.addEventListener("DOMContentLoaded", async () => {
  const burger = document.querySelector(".burger");
  const nav = document.querySelector(".nav-links");
  const dropdown = document.querySelector(".dropdown");
  const dropbtn = document.querySelector(".dropbtn");

  // 🔹 Vérifier si l'utilisateur est admin et cacher le bouton si nécessaire
  try {
    const response = await fetch("/api/check-admin");
    const data = await response.json();

    const adminLink = document.querySelector(
      'a[href="Admin"], a[href="/Admin"]'
    );
    if (adminLink && !data.isAdmin) {
      // Cacher le li parent du lien Admin
      adminLink.closest("li").style.display = "none";
    }
  } catch (error) {
    console.error("Erreur lors de la vérification des droits admin:", error);
    // En cas d'erreur, on cache le bouton par sécurité
    const adminLink = document.querySelector(
      'a[href="Admin"], a[href="/Admin"]'
    );
    if (adminLink) {
      adminLink.closest("li").style.display = "none";
    }
  }

  console.log(
    "Nombre de boutons trouvés :",
    document.querySelectorAll(".nav-links li a").length
  );

  // Burger menu
  if (burger && nav) {
    burger.addEventListener("click", () => {
      nav.classList.toggle("nav-active");
      burger.classList.toggle("toggle");
    });
  }

  // Dropdown
  if (dropbtn && dropdown) {
    dropbtn.addEventListener("click", (e) => {
      e.preventDefault();
      dropdown.classList.toggle("active");
    });

    document.addEventListener("click", (e) => {
      if (!dropdown.contains(e.target) && !dropbtn.contains(e.target)) {
        dropdown.classList.remove("active");
      }
    });
  }

  // Fermer le menu mobile au clic sur un lien
  document.querySelectorAll(".nav-links li a").forEach((link) => {
    link.addEventListener("click", (e) => {
      if (!link.classList.contains("dropbtn")) {
        nav.classList.remove("nav-active");
        burger.classList.remove("toggle");
      }
    });
  });
});

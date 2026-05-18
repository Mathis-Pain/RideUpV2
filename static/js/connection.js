document.addEventListener("DOMContentLoaded", () => {
  const banner = document.getElementById("cookie-banner");
  const hasConsent = document.cookie.includes("cookieConsent=");

  if (!hasConsent) {
    banner.classList.remove("hidden");
    setTimeout(() => banner.classList.add("show"), 50);
  }

  document.getElementById("accept-cookies").addEventListener("click", () => {
    document.cookie =
      "cookieConsent=accepted; path=/; max-age=" + 60 * 60 * 24 * 365;
    hideBanner();
  });

  document.getElementById("decline-cookies").addEventListener("click", () => {
    document.cookie =
      "cookieConsent=declined; path=/; max-age=" + 60 * 60 * 24 * 365;
    hideBanner();
    // Tu peux désactiver ici Google Analytics ou autres scripts
  });

  function hideBanner() {
    banner.classList.remove("show");
    setTimeout(() => banner.classList.add("hidden"), 500);
  }
});
function togglePassword(fieldId, checkbox) {
  const passwordField = document.getElementById(fieldId);
  if (passwordField) {
    if (checkbox.checked) {
      passwordField.type = "text";
    } else {
      passwordField.type = "password";
    }
  }
}

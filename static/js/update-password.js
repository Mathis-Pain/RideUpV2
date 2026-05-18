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

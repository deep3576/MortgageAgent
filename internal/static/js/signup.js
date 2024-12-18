document.addEventListener('DOMContentLoaded', function() {
    const passwordField = document.querySelector('#password');
    const closedEyeIcon = document.querySelector('#closedEyeIcon');
    const openEyeIcon = document.querySelector('#openEyeIcon');
    
    const confirmPasswordField = document.querySelector('#confirm_password');
    const closedEyeIconConfirm = document.querySelector('#closedEyeIconConfirm');
    const openEyeIconConfirm = document.querySelector('#openEyeIconConfirm');

    // Toggle visibility for main password
    closedEyeIcon.addEventListener('click', () => {
        passwordField.setAttribute('type', 'text');
        closedEyeIcon.style.display = 'none';
        openEyeIcon.style.display = 'block';
    });

    openEyeIcon.addEventListener('click', () => {
        passwordField.setAttribute('type', 'password');
        openEyeIcon.style.display = 'none';
        closedEyeIcon.style.display = 'block';
    });

    // Toggle visibility for confirm password
    closedEyeIconConfirm.addEventListener('click', () => {
        confirmPasswordField.setAttribute('type', 'text');
        closedEyeIconConfirm.style.display = 'none';
        openEyeIconConfirm.style.display = 'block';
    });

    openEyeIconConfirm.addEventListener('click', () => {
        confirmPasswordField.setAttribute('type', 'password');
        openEyeIconConfirm.style.display = 'none';
        closedEyeIconConfirm.style.display = 'block';
    });
});

function validatePasswords() {
    const pwd = document.getElementById('password').value;
    const confirmPwd = document.getElementById('confirm_password').value;
    const errorMessage = document.getElementById('error-message');

    if (pwd !== confirmPwd) {
        errorMessage.style.display = 'block';
        return false; // prevent form submission
    }
    errorMessage.style.display = 'none';
    return true; // allow form submission
}

document.addEventListener('DOMContentLoaded', function() {
    // 1. Scroll-based reveal animations
    const revealElements = document.querySelectorAll('.reveal-on-scroll');

    function animateOnScroll() {
        const windowHeight = window.innerHeight;
        const revealPoint = 150; // when element is 150px into the viewport
        
        for (let i = 0; i < revealElements.length; i++) {
            const element = revealElements[i];
            const elementTop = element.getBoundingClientRect().top;

            if (elementTop < windowHeight - revealPoint) {
                element.classList.add('active');
            } else {
                element.classList.remove('active');
            }
        }
    }

    window.addEventListener('scroll', animateOnScroll);
    animateOnScroll(); // Initial check on page load

    // 2. Input field interaction animation
    const inputFields = document.querySelectorAll('input[type="text"], input[type="password"], input[type="email"]');

    inputFields.forEach((input) => {
        input.addEventListener('focus', () => {
            input.classList.add('input-focus-animation');
        });
        input.addEventListener('blur', () => {
            input.classList.remove('input-focus-animation');
        });
    });
});

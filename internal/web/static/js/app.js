document.addEventListener("DOMContentLoaded", () => {
    const closeButtons = document.querySelectorAll("[data-close]");

    closeButtons.forEach((button) => {
        button.addEventListener("click", () => {
            const target = document.querySelector(button.dataset.close);

            if (target) {
                target.hidden = true;
            }
        });
    });
});
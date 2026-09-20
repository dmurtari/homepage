const $ = document.querySelector.bind(document);
const SCROLL_AMOUNT = 300;
const SCROLL_TOLERANCE = 10;

function updateNavButtons() {
    const nav = $("#slideshow-nav");
    $("#scroll-left-button").disabled = nav.scrollLeft <= SCROLL_TOLERANCE;
    $("#scroll-right-button").disabled =
        nav.scrollLeft + nav.clientWidth >= nav.scrollWidth - SCROLL_TOLERANCE;
}

function scrollActiveNavItemIntoView() {
    const activeNavItem = document.querySelector(".slideshow__nav-item.active");

    if (activeNavItem) {
        activeNavItem.scrollIntoView({
            block: "nearest",
            inline: "nearest",
        });
    }
}

function scrollNavLeft() {
    const nav = $("#slideshow-nav");
    nav.scrollBy({ left: -SCROLL_AMOUNT, behavior: "smooth" });
}

function scrollNavRight() {
    const nav = $("#slideshow-nav");
    nav.scrollBy({ left: SCROLL_AMOUNT, behavior: "smooth" });
}

document.addEventListener(
    "scroll",
    (e) => {
        if (e.target.id === "slideshow-nav") {
            updateNavButtons();
        }
    },
    true,
);

document.addEventListener("load", updateNavButtons, true);

document.addEventListener("click", (e) => {
    if (e.target.id === "scroll-left-button") {
        scrollNavLeft();
    } else if (e.target.id === "scroll-right-button") {
        scrollNavRight();
    }
});

document.addEventListener("htmx:after:swap", scrollActiveNavItemIntoView);
window.addEventListener("load", scrollActiveNavItemIntoView);

// Show an element
var showSection = function(elem, arrow) {
    // Get the natural height of the element
    var getHeight = function() {
        elem.style.display = 'block'; // Make it visible
        var height = elem.scrollHeight + 'px'; // Get it's height
        elem.style.display = ''; //  Hide it again
        return height;
    };

    var height = getHeight(); // Get the natural height
    elem.classList.add('is-visible'); // Make the element visible
    elem.style.height = height; // Update the max-height

    // Once the transition is complete, remove the inline max-height so the content can scale responsively
    window.setTimeout(function() {
        elem.style.height = '';
    }, 350);

    if (arrow) {
        arrow.style.transform = "rotate(0deg)";
    }

};

// Hide an element
var hideSection = function(elem, arrow) {
    // Give the element a height to change from
    elem.style.height = elem.scrollHeight + 'px';

    // Set the height back to 0
    window.setTimeout(function() {
        elem.style.height = '0';
    }, 1);

    // When the transition is complete, hide it
    window.setTimeout(function() {
        elem.classList.remove('is-visible');
    }, 350);

    if (arrow) {
        arrow.style.transform = "rotate(-90deg)";
    }

};

// Toggle element visibility
var toggleSection = function(elem, arrow) {

    // If the element is visible, hide it
    if (elem.classList.contains('is-visible')) {
        hideSection(elem, arrow);
        return;
    }

    // Otherwise, show it
    showSection(elem, arrow);

};

// Listen for click events
document.addEventListener('click', function(event) {
    // Make sure clicked element is our toggle or the arrow of the toggle
    let target = null;
    if (event.target.tagName == "svg") {
        if (event.target.parentElement.classList.contains("arrow")) {
            target = event.target.parentElement.parentElement;
        }
    }
    if (!target) {
        if (!event.target.classList.contains('toggle')) return;
        target = event.target;
    }

    // Prevent default link behavior
    event.preventDefault();

    // Get the content
    var content = document.querySelector(target.hash);
    if (!content) return;

    // Toggle the content
    toggleSection(content, target.querySelector(".arrow"));

}, false);


/* Toggle mobile menu */
function toggleMenu() {
    const menu = document.querySelector(".menu");
    if (menu.classList.contains("active")) {
        menu.classList.remove("active");
        toggleSection.querySelector("a").innerHTML = "<i class='fas fa-bars'></i>";
    } else {
        menu.classList.add("active");
        toggleSection.querySelector("a").innerHTML = "<i class='fas fa-times'></i>";
    }
}

/* Activate Submenu */
function toggleItem() {
    const menu = document.querySelector(".menu");
    if (this.classList.contains("submenu-active")) {
        this.classList.remove("submenu-active");
    } else if (menu.querySelector(".submenu-active")) {
        menu.querySelector(".submenu-active").classList.remove("submenu-active");
        this.classList.add("submenu-active");
    } else {
        this.classList.add("submenu-active");
    }
}

/* Close Submenu From Anywhere */
function closeSubmenu(e) {
    const menu = document.querySelector(".menu");
    if (menu.querySelector(".submenu-active")) {
        let isClickInside = menu
            .querySelector(".submenu-active")
            .contains(e.target);

        if (!isClickInside && menu.querySelector(".submenu-active")) {
            menu.querySelector(".submenu-active").classList.remove("submenu-active");
        }
    }
}

document.addEventListener("DOMContentLoaded", function() {

    document.querySelector(".toggle-menu").addEventListener("click", toggleMenu, false);

    for (let item of document.querySelectorAll(".item")) {
        if (item.querySelector(".submenu")) {
            item.addEventListener("click", toggleItem, false);
        }
        item.addEventListener("keypress", toggleItem, false);
    }
    document.addEventListener("click", closeSubmenu, false);
});
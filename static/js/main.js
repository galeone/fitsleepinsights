// Show an element
var show = function (elem, arrow) {

	// Get the natural height of the element
	var getHeight = function () {
		elem.style.display = 'block'; // Make it visible
		var height = elem.scrollHeight + 'px'; // Get it's height
		elem.style.display = ''; //  Hide it again
		return height;
	};

	var height = getHeight(); // Get the natural height
	elem.classList.add('is-visible'); // Make the element visible
	elem.style.height = height; // Update the max-height

	// Once the transition is complete, remove the inline max-height so the content can scale responsively
	window.setTimeout(function () {
		elem.style.height = '';
	}, 350);

    if (arrow) {
        arrow.style.transform = "rotate(0deg)";
    }

};

// Hide an element
var hide = function (elem, arrow) {

	// Give the element a height to change from
	elem.style.height = elem.scrollHeight + 'px';

	// Set the height back to 0
	window.setTimeout(function () {
		elem.style.height = '0';
	}, 1);

	// When the transition is complete, hide it
	window.setTimeout(function () {
		elem.classList.remove('is-visible');
	}, 350);

    if (arrow) {
        arrow.style.transform = "rotate(-90deg)";
    }

};

// Toggle element visibility
var toggle = function (elem, arrow) {

	// If the element is visible, hide it
	if (elem.classList.contains('is-visible')) {
		hide(elem, arrow);
		return;
	}

	// Otherwise, show it
	show(elem, arrow);
	
};

// Listen for click events
document.addEventListener('click', function (event) {

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
	toggle(content, target.querySelector(".arrow"));

}, false);

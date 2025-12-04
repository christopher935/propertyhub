// Quick View Modal - Pattern 3: Hybrid Quick View
// Handles modal open/close, image gallery, and property data display

// Sample property data (will be replaced with real data later)
const sampleProperty = {
    id: 1,
    address: "1426 Ashwood Dr",
    city: "Houston",
    state: "TX",
    zip: "77077",
    price: 2500,
    beds: 3,
    baths: 2,
    sqft: 1650,
    images: [
        "https://images.unsplash.com/photo-1568605114967-8130f3a36994?w=1200&h=800&fit=crop",
        "https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?w=1200&h=800&fit=crop",
        "https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?w=1200&h=800&fit=crop",
        "https://images.unsplash.com/photo-1600566753190-17f0baa2a6c3?w=1200&h=800&fit=crop"
    ],
    description: "Beautiful 3-bedroom, 2-bathroom home in a quiet neighborhood. This well-maintained property features an open floor plan, modern kitchen with stainless steel appliances, spacious master suite, and a large backyard perfect for entertaining.",
    yearBuilt: 2015,
    propertyType: "Single Family",
    parking: "2-Car Garage",
    yard: "Large Yard"
};

// Current image index for gallery
let currentImageIndex = 0;

// Open Quick View Modal
function openQuickView(propertyId) {
    // In production, fetch property data by ID
    // For now, use sample data
    const property = sampleProperty;
    
    // Populate modal with property data
    document.getElementById('qv-address').textContent = property.address;
    document.getElementById('qv-location').textContent = `${property.city}, ${property.state} ${property.zip}`;
    document.getElementById('qv-price').textContent = `$${property.price.toLocaleString()}`;
    document.getElementById('qv-beds').innerHTML = `<strong>${property.beds}</strong> Beds`;
    document.getElementById('qv-baths').innerHTML = `<strong>${property.baths}</strong> Baths`;
    document.getElementById('qv-sqft').innerHTML = `<strong>${property.sqft.toLocaleString()}</strong> sq ft`;
    document.getElementById('qv-description').textContent = property.description;
    document.getElementById('qv-full-details-link').href = `/property/${property.id}`;
    
    // Set up image gallery
    currentImageIndex = 0;
    updateGallery(property.images);
    
    // Show modal with animation
    const modal = document.getElementById('quick-view-modal');
    modal.style.display = 'flex';
    // Trigger reflow for animation
    modal.offsetHeight;
    modal.classList.add('active');
    
    // Prevent body scroll
    document.body.style.overflow = 'hidden';
}

// Close Quick View Modal
function closeQuickView() {
    const modal = document.getElementById('quick-view-modal');
    modal.classList.remove('active');
    
    // Wait for animation to complete before hiding
    setTimeout(() => {
        modal.style.display = 'none';
        document.body.style.overflow = '';
    }, 300);
}

// Update Gallery Images
function updateGallery(images) {
    // Update main image
    document.getElementById('qv-main-image').src = images[currentImageIndex];
    
    // Update counter
    document.getElementById('qv-image-counter').textContent = `${currentImageIndex + 1} / ${images.length}`;
    
    // Update thumbnails
    const thumbnailsContainer = document.getElementById('qv-thumbnails');
    thumbnailsContainer.innerHTML = '';
    
    images.forEach((img, index) => {
        const thumb = document.createElement('div');
        thumb.className = `gallery-thumbnail ${index === currentImageIndex ? 'active' : ''}`;
        thumb.onclick = () => {
            currentImageIndex = index;
            updateGallery(images);
        };
        
        const thumbImg = document.createElement('img');
        thumbImg.src = img;
        thumbImg.alt = `Thumbnail ${index + 1}`;
        
        thumb.appendChild(thumbImg);
        thumbnailsContainer.appendChild(thumb);
    });
}

// Navigate to previous image
function prevImage() {
    const images = sampleProperty.images;
    currentImageIndex = (currentImageIndex - 1 + images.length) % images.length;
    updateGallery(images);
}

// Navigate to next image
function nextImage() {
    const images = sampleProperty.images;
    currentImageIndex = (currentImageIndex + 1) % images.length;
    updateGallery(images);
}

// Close modal on ESC key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        closeQuickView();
    }
});

// Close modal on backdrop click
document.addEventListener('click', (e) => {
    const modal = document.getElementById('quick-view-modal');
    if (e.target === modal) {
        closeQuickView();
    }
});

// Prevent modal content clicks from closing modal
document.addEventListener('DOMContentLoaded', () => {
    const modalContainer = document.querySelector('.modal-container');
    if (modalContainer) {
        modalContainer.addEventListener('click', (e) => {
            e.stopPropagation();
        });
    }
});

// Quick View Modal - Pattern 3: Hybrid Quick View
// Handles modal open/close, image gallery, and property data display

// Current image index for gallery
let currentImageIndex = 0;
// Current property data (loaded from API)
let currentProperty = null;

// Open Quick View Modal
async function openQuickView(propertyId) {
    const modal = document.getElementById('quick-view-modal');
    
    // Show modal with loading state
    modal.style.display = 'flex';
    modal.offsetHeight;
    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
    
    // Show loading state in modal
    const modalContent = modal.querySelector('.quick-view-content');
    if (modalContent) {
        modalContent.classList.add('loading');
    }
    
    try {
        const response = await fetch(`/api/properties/${propertyId}`);
        if (!response.ok) {
            throw new Error('Failed to load property');
        }
        
        const result = await response.json();
        if (!result.success || !result.data) {
            throw new Error('Invalid response format');
        }
        
        const property = result.data;
        currentProperty = property;
        
        // Populate modal with property data
        document.getElementById('qv-address').textContent = property.address || 'Address not available';
        document.getElementById('qv-location').textContent = `${property.city || ''}, ${property.state || ''} ${property.zip_code || ''}`;
        document.getElementById('qv-price').textContent = `$${(property.price || 0).toLocaleString()}`;
        document.getElementById('qv-beds').innerHTML = `<strong>${property.bedrooms || 0}</strong> Beds`;
        document.getElementById('qv-baths').innerHTML = `<strong>${property.bathrooms || 0}</strong> Baths`;
        document.getElementById('qv-sqft').innerHTML = `<strong>${(property.square_feet || 0).toLocaleString()}</strong> sq ft`;
        document.getElementById('qv-description').textContent = property.description || 'No description available';
        document.getElementById('qv-full-details-link').href = `/property/${property.id}`;
        
        // Set up image gallery
        currentImageIndex = 0;
        const images = property.images && property.images.length > 0 ? property.images : [];
        updateGallery(images);
        
        // Remove loading state
        if (modalContent) {
            modalContent.classList.remove('loading');
        }
    } catch (error) {
        console.error('Quick view error:', error);
        
        // Show error state
        const errorHtml = `
            <div style="text-align: center; padding: 40px;">
                <p style="color: #ef4444; font-size: 18px; margin-bottom: 10px;">Failed to load property details</p>
                <p style="color: #6b7280;">Please try again later</p>
            </div>
        `;
        
        if (modalContent) {
            modalContent.innerHTML = errorHtml;
            modalContent.classList.remove('loading');
        }
    }
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
    // Handle empty images array
    if (!images || images.length === 0) {
        const mainImageContainer = document.getElementById('qv-main-image');
        if (mainImageContainer) {
            mainImageContainer.src = 'https://via.placeholder.com/800x600?text=No+Image+Available';
        }
        const counter = document.getElementById('qv-image-counter');
        if (counter) {
            counter.textContent = '0 / 0';
        }
        const thumbnailsContainer = document.getElementById('qv-thumbnails');
        if (thumbnailsContainer) {
            thumbnailsContainer.innerHTML = '';
        }
        return;
    }
    
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
    if (!currentProperty || !currentProperty.images || currentProperty.images.length === 0) {
        return;
    }
    const images = currentProperty.images;
    currentImageIndex = (currentImageIndex - 1 + images.length) % images.length;
    updateGallery(images);
}

// Navigate to next image
function nextImage() {
    if (!currentProperty || !currentProperty.images || currentProperty.images.length === 0) {
        return;
    }
    const images = currentProperty.images;
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

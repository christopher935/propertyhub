// Homepage Properties Alpine.js Component
// Avoids Go template parsing conflicts by keeping JavaScript in external file

function homepageProperties() {
    return {
        properties: [],
        loading: true,
        error: null,
        page: 1,
        limit: 12,
        hasMore: true,
        selectedProperty: null,
        showModal: false,
        currentImageIndex: 0,
        
        async init() {
            await this.loadProperties();
            // Setup infinite scroll
            window.addEventListener('scroll', () => {
                if (this.shouldLoadMore()) {
                    this.loadMore();
                }
            });
        },
        
        async loadProperties() {
            try {
                this.loading = true;
                const response = await fetch(`/api/properties?page=${this.page}&limit=${this.limit}`);
                const data = await response.json();
                
                if (data.success && data.data) {
                    const properties = Array.isArray(data.data) ? data.data : (data.data.properties || []);
                    if (this.page === 1) {
                        this.properties = properties;
                    } else {
                        this.properties = [...this.properties, ...properties];
                    }
                    
                    this.hasMore = properties.length === this.limit;
                } else {
                    this.error = 'Failed to load properties';
                }
            } catch (err) {
                console.error('Error fetching properties:', err);
                this.error = 'Failed to load properties';
            } finally {
                this.loading = false;
            }
        },
        
        shouldLoadMore() {
            if (!this.hasMore || this.loading) return false;
            const scrollPosition = window.innerHeight + window.scrollY;
            const threshold = document.documentElement.scrollHeight - 500;
            return scrollPosition >= threshold;
        },
        
        async loadMore() {
            this.page++;
            await this.loadProperties();
        },
        
        async openQuickView(propertyId) {
            try {
                const response = await fetch(`/api/properties/${propertyId}`);
                const data = await response.json();
                
                if (data.success && data.data) {
                    this.selectedProperty = Array.isArray(data.data) ? data.data[0] : (data.data.property || data.data);
                    this.currentImageIndex = 0;
                    this.showModal = true;
                    
                    const modal = document.getElementById('quick-view-modal');
                    const backdrop = modal.querySelector('.modal-backdrop');
                    const wrapper = modal.querySelector('.modal-wrapper');
                    
                    modal.style.display = 'flex';
                    document.body.style.overflow = 'hidden';
                    
                    // Trigger animation
                    requestAnimationFrame(() => {
                        if (backdrop) backdrop.classList.add('show');
                        if (wrapper) wrapper.classList.add('show');
                    });
                    
                    // Update modal content
                    this.updateModalContent();
                }
            } catch (err) {
                console.error('Error fetching property details:', err);
                alert('Failed to load property details');
            }
        },
        
        updateModalContent() {
            if (!this.selectedProperty) return;
            
            const p = this.selectedProperty;
            document.getElementById('qv-address').textContent = this.formatAddress(p.address);
            document.getElementById('qv-location').textContent = `${p.city || 'Houston'}, TX ${p.zip || ''}`;
            document.getElementById('qv-price').textContent = '$' + (p.price || 0).toLocaleString();
            document.getElementById('qv-beds').innerHTML = `<strong>${p.bedrooms || 0}</strong> Beds`;
            document.getElementById('qv-baths').innerHTML = `<strong>${p.bathrooms || 0}</strong> Baths`;
            document.getElementById('qv-sqft').innerHTML = `<strong>${(p.square_feet || 0).toLocaleString()}</strong> sq ft`;
            document.getElementById('qv-description').textContent = p.description || 'No description available.';
            document.getElementById('qv-full-details-link').href = '/properties/' + p.id;
            
            // Update gallery if images exist
            let imagesToShow = [];
            if (p.featured_image) {
                imagesToShow = [p.featured_image];
                if (p.images && p.images.length > 0) {
                    imagesToShow = imagesToShow.concat(p.images.filter(img => img !== p.featured_image));
                }
            } else if (p.images && p.images.length > 0) {
                imagesToShow = p.images;
            }
            
            if (imagesToShow.length > 0) {
                this.updateGallery(imagesToShow);
            } else {
                const gallery = document.querySelector('.gallery-main');
                if (gallery) gallery.style.display = 'none';
            }
        },
        
        updateGallery(images) {
            if (!images || images.length === 0) {
                const gallery = document.querySelector('.gallery-main');
                if (gallery) gallery.style.display = 'none';
                return;
            }
            
            const mainImage = document.getElementById('qv-main-image');
            const counter = document.getElementById('qv-image-counter');
            const thumbnails = document.getElementById('qv-thumbnails');
            const gallery = document.querySelector('.gallery-main');
            
            if (mainImage) {
                mainImage.src = images[this.currentImageIndex];
                mainImage.onerror = () => {
                    if (gallery) gallery.style.display = 'none';
                };
                if (gallery) gallery.style.display = 'block';
            }
            
            if (counter) {
                counter.textContent = `${this.currentImageIndex + 1} / ${images.length}`;
            }
            
            if (thumbnails) {
                thumbnails.innerHTML = '';
                images.forEach((img, index) => {
                    const thumb = document.createElement('div');
                    thumb.className = 'qv-thumbnail' + (index === this.currentImageIndex ? ' active' : '');
                    thumb.onclick = () => {
                        this.currentImageIndex = index;
                        this.updateGallery(images);
                    };
                    
                    const thumbImg = document.createElement('img');
                    thumbImg.src = img;
                    thumbImg.alt = 'Thumbnail ' + (index + 1);
                    thumbImg.onerror = () => {
                        thumb.style.display = 'none';
                    };
                    
                    thumb.appendChild(thumbImg);
                    thumbnails.appendChild(thumb);
                });
            }
        },
        
        prevImage() {
            if (!this.selectedProperty || !this.selectedProperty.images) return;
            const images = this.selectedProperty.images;
            this.currentImageIndex = (this.currentImageIndex - 1 + images.length) % images.length;
            this.updateGallery(images);
        },
        
        nextImage() {
            if (!this.selectedProperty || !this.selectedProperty.images) return;
            const images = this.selectedProperty.images;
            this.currentImageIndex = (this.currentImageIndex + 1) % images.length;
            this.updateGallery(images);
        },
        
        closeQuickView() {
            const modal = document.getElementById('quick-view-modal');
            const backdrop = modal ? modal.querySelector('.modal-backdrop') : null;
            const wrapper = modal ? modal.querySelector('.modal-wrapper') : null;
            
            if (backdrop) backdrop.classList.remove('show');
            if (wrapper) wrapper.classList.remove('show');
            
            setTimeout(() => {
                this.showModal = false;
                this.selectedProperty = null;
                if (modal) modal.style.display = 'none';
                document.body.style.overflow = '';
            }, 200);
        },
        
        formatPrice(price) {
            return new Intl.NumberFormat('en-US', {
                style: 'currency',
                currency: 'USD',
                minimumFractionDigits: 0
            }).format(price);
        },
        
        formatAddress(fullAddress) {
            if (!fullAddress) return '';
            // Extract just street number and name (before first comma)
            const parts = fullAddress.split(',');
            return parts[0].trim();
        }
    };
}

/**
 * PropertyHub Google Maps Integration
 * Advanced maps with property markers, neighborhoods, search, routing
 * Real estate-focused mapping with MLS integration and location intelligence
 */

class PropertyHubMaps {
    constructor() {
        this.maps = new Map();
        this.markers = new Map();
        this.infoWindows = new Map();
        this.clusterers = new Map();
        this.drawingManagers = new Map();
        this.directionsService = null;
        this.directionsRenderer = null;
        this.geocoder = null;
        this.placesService = null;
        this.streetViewService = null;
        
        // Map configurations
        this.defaultConfig = {
            zoom: 12,
            center: { lat: 29.7604, lng: -95.3698 }, // Houston, TX
            mapTypeId: 'roadmap',
            styles: this.getMapStyles(),
            disableDefaultUI: false,
            zoomControl: true,
            mapTypeControl: true,
            scaleControl: true,
            streetViewControl: true,
            rotateControl: false,
            fullscreenControl: true
        };

        // Property marker configurations
        this.markerConfig = {
            default: {
                icon: '/static/images/markers/property-default.png',
                scaledSize: { width: 32, height: 32 }
            },
            available: {
                icon: '/static/images/markers/property-available.png',
                scaledSize: { width: 32, height: 32 }
            },
            sold: {
                icon: '/static/images/markers/property-sold.png',
                scaledSize: { width: 32, height: 32 }
            },
            pending: {
                icon: '/static/images/markers/property-pending.png',
                scaledSize: { width: 32, height: 32 }
            },
            featured: {
                icon: '/static/images/markers/property-featured.png',
                scaledSize: { width: 40, height: 40 }
            }
        };

        // Neighborhood overlays
        this.neighborhoods = new Map();
        this.schoolDistricts = new Map();
        this.transitLines = new Map();

        this.init();
    }

    async init() {
        console.log('🗺️ PropertyHub Maps initializing...');
        
        await this.loadGoogleMapsAPI();
        this.setupMapInstances();
        this.initializeServices();
        this.setupEventListeners();
        
        console.log('✅ PropertyHub Maps ready');
    }

    // Google Maps API Loading
    async loadGoogleMapsAPI() {
        if (typeof google !== 'undefined' && google.maps) {
            return Promise.resolve();
        }

        return new Promise((resolve, reject) => {
            // Check if script is already loading
            if (window.googleMapsLoading) {
                window.googleMapsLoading.then(resolve).catch(reject);
                return;
            }

            const apiKey = window.GOOGLE_MAPS_API_KEY || 'your-api-key-here';
            const script = document.createElement('script');
            
            script.src = `https://maps.googleapis.com/maps/api/js?key=${apiKey}&libraries=places,geometry,drawing`;
            script.async = true;
            script.defer = true;
            
            window.googleMapsLoading = new Promise((scriptResolve, scriptReject) => {
                script.onload = () => {
                    delete window.googleMapsLoading;
                    scriptResolve();
                };
                script.onerror = () => {
                    delete window.googleMapsLoading;
                    scriptReject(new Error('Failed to load Google Maps API'));
                };
            });

            document.head.appendChild(script);
            
            window.googleMapsLoading.then(resolve).catch(reject);
        });
    }

    // Map Instance Management
    setupMapInstances() {
        const mapElements = document.querySelectorAll('.property-map');
        
        mapElements.forEach(element => {
            this.initializeMap(element);
        });
    }

    initializeMap(element) {
        const mapId = element.id || this.generateMapId();
        const mapType = element.dataset.mapType || 'property';
        const propertyId = element.dataset.propertyId;
        const showControls = element.dataset.controls !== 'false';
        const showCluster = element.dataset.cluster === 'true';
        const enableDrawing = element.dataset.drawing === 'true';

        // Get initial configuration
        const config = {
            ...this.defaultConfig,
            disableDefaultUI: !showControls
        };

        // Override center if specified
        if (element.dataset.lat && element.dataset.lng) {
            config.center = {
                lat: parseFloat(element.dataset.lat),
                lng: parseFloat(element.dataset.lng)
            };
        }

        // Override zoom if specified
        if (element.dataset.zoom) {
            config.zoom = parseInt(element.dataset.zoom);
        }

        // Create map instance
        const mapInstance = new google.maps.Map(element, config);

        // Store map configuration
        const mapConfig = {
            element: element,
            instance: mapInstance,
            id: mapId,
            type: mapType,
            propertyId: propertyId,
            showCluster: showCluster,
            enableDrawing: enableDrawing,
            markers: [],
            infoWindow: null,
            clusterer: null,
            drawingManager: null
        };

        this.maps.set(mapId, mapConfig);

        // Setup map-specific features
        this.setupMapFeatures(mapId, mapConfig);
    }

    async setupMapFeatures(mapId, config) {
        // Setup clustering if enabled
        if (config.showCluster) {
            this.setupMarkerClustering(mapId);
        }

        // Setup drawing tools if enabled
        if (config.enableDrawing) {
            this.setupDrawingTools(mapId);
        }

        // Setup info window
        config.infoWindow = new google.maps.InfoWindow();

        // Load map-specific data
        switch (config.type) {
            case 'property':
                await this.loadPropertyData(mapId, config.propertyId);
                break;
            case 'search':
                await this.loadSearchProperties(mapId);
                break;
            case 'neighborhood':
                await this.loadNeighborhoodData(mapId);
                break;
        }

        // Setup map event listeners
        this.setupMapEventListeners(mapId, config);
    }

    initializeServices() {
        if (typeof google !== 'undefined' && google.maps) {
            this.directionsService = new google.maps.DirectionsService();
            this.directionsRenderer = new google.maps.DirectionsRenderer();
            this.geocoder = new google.maps.Geocoder();
            this.streetViewService = new google.maps.StreetViewService();
        }
    }

    // Property Markers
    async loadPropertyData(mapId, propertyId) {
        try {
            let url = '/api/properties/map-data';
            if (propertyId) {
                url += `?propertyId=${propertyId}`;
            }

            const response = await fetch(url, {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to load property data');
            }

            const data = await response.json();
            this.addPropertyMarkers(mapId, data.properties);

        } catch (error) {
            console.error('Error loading property data:', error);
        }
    }

    addPropertyMarkers(mapId, properties) {
        const config = this.maps.get(mapId);
        if (!config) return;

        const bounds = new google.maps.LatLngBounds();
        const markers = [];

        properties.forEach(property => {
            if (property.latitude && property.longitude) {
                const marker = this.createPropertyMarker(mapId, property);
                markers.push(marker);
                bounds.extend(marker.getPosition());
            }
        });

        config.markers = markers;

        // Setup clustering if enabled
        if (config.showCluster && config.clusterer) {
            config.clusterer.addMarkers(markers);
        }

        // Fit map to show all markers
        if (markers.length > 1) {
            config.instance.fitBounds(bounds);
        } else if (markers.length === 1) {
            config.instance.setCenter(markers[0].getPosition());
            config.instance.setZoom(15);
        }
    }

    createPropertyMarker(mapId, property) {
        const config = this.maps.get(mapId);
        const position = {
            lat: parseFloat(property.latitude),
            lng: parseFloat(property.longitude)
        };

        // Select appropriate marker icon
        let markerType = 'default';
        if (property.featured) markerType = 'featured';
        else if (property.status === 'available') markerType = 'available';
        else if (property.status === 'sold') markerType = 'sold';
        else if (property.status === 'pending') markerType = 'pending';

        const markerConfig = this.markerConfig[markerType];
        
        const marker = new google.maps.Marker({
            position: position,
            map: config.instance,
            title: property.address,
            icon: {
                url: markerConfig.icon,
                scaledSize: new google.maps.Size(
                    markerConfig.scaledSize.width,
                    markerConfig.scaledSize.height
                )
            },
            property: property
        });

        // Add click listener for info window
        marker.addListener('click', () => {
            this.showPropertyInfoWindow(mapId, marker, property);
        });

        // Add hover listeners
        marker.addListener('mouseover', () => {
            marker.setZIndex(1000);
            this.showPropertyTooltip(mapId, marker, property);
        });

        marker.addListener('mouseout', () => {
            marker.setZIndex(1);
            this.hidePropertyTooltip(mapId);
        });

        return marker;
    }

    showPropertyInfoWindow(mapId, marker, property) {
        const config = this.maps.get(mapId);
        const infoWindow = config.infoWindow;
        
        const content = this.createPropertyInfoWindowContent(property);
        infoWindow.setContent(content);
        infoWindow.open(config.instance, marker);

        // Track analytics
        this.trackMapInteraction('property_info_viewed', {
            propertyId: property.id,
            address: property.address
        });
    }

    createPropertyInfoWindowContent(property) {
        const mainPhoto = property.photos && property.photos.length > 0 
            ? property.photos[0].url 
            : '/static/images/property-placeholder.jpg';

        return `
            <div class="map-info-window">
                <div class="info-image">
                    <img src="${mainPhoto}" alt="${property.address}" 
                         style="width: 250px; height: 150px; object-fit: cover;">
                    ${property.featured ? '<div class="featured-badge">Featured</div>' : ''}
                </div>
                
                <div class="info-content">
                    <div class="info-price">$${property.price.toLocaleString()}</div>
                    <div class="info-address">${property.address}</div>
                    <div class="info-city">${property.city}, ${property.state} ${property.zipCode}</div>
                    
                    <div class="info-details">
                        <span class="detail-item">
                            <i class="fas fa-bed"></i> ${property.bedrooms} beds
                        </span>
                        <span class="detail-item">
                            <i class="fas fa-bath"></i> ${property.bathrooms} baths
                        </span>
                        <span class="detail-item">
                            <i class="fas fa-ruler-combined"></i> ${property.squareFootage?.toLocaleString()} sqft
                        </span>
                    </div>
                    
                    ${property.description ? `
                        <div class="info-description">
                            ${property.description.substring(0, 120)}...
                        </div>
                    ` : ''}
                    
                    <div class="info-actions">
                        <a href="/properties/${property.id}" class="btn btn-primary btn-sm">
                            View Details
                        </a>
                        <button class="btn btn-secondary btn-sm" 
                                onclick="PropertyHubMaps.getDirections('${property.address}')">
                            Directions
                        </button>
                        <button class="btn btn-outline btn-sm" 
                                onclick="PropertyHubMaps.showStreetView(${property.latitude}, ${property.longitude})">
                            Street View
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    // Marker Clustering
    setupMarkerClustering(mapId) {
        const config = this.maps.get(mapId);
        
        // Using MarkerClusterer (would need to include the library)
        if (typeof MarkerClusterer !== 'undefined') {
            config.clusterer = new MarkerClusterer(config.instance, [], {
                styles: [{
                    textColor: 'white',
                    url: '/static/images/cluster-icon.png',
                    height: 40,
                    width: 40
                }],
                maxZoom: 15
            });
        }
    }

    // Drawing Tools
    setupDrawingTools(mapId) {
        const config = this.maps.get(mapId);
        
        const drawingManager = new google.maps.drawing.DrawingManager({
            drawingMode: null,
            drawingControl: true,
            drawingControlOptions: {
                position: google.maps.ControlPosition.TOP_CENTER,
                drawingModes: [
                    google.maps.drawing.OverlayType.POLYGON,
                    google.maps.drawing.OverlayType.RECTANGLE,
                    google.maps.drawing.OverlayType.CIRCLE
                ]
            },
            polygonOptions: {
                fillColor: '#3b82f6',
                fillOpacity: 0.3,
                strokeWeight: 2,
                strokeColor: '#1d4ed8',
                clickable: false,
                editable: true,
                zIndex: 1
            }
        });

        drawingManager.setMap(config.instance);
        config.drawingManager = drawingManager;

        // Listen for drawing completion
        google.maps.event.addListener(drawingManager, 'overlaycomplete', (event) => {
            this.handleDrawingComplete(mapId, event);
        });
    }

    handleDrawingComplete(mapId, event) {
        const overlay = event.overlay;
        const type = event.type;
        
        let coordinates = [];
        
        switch (type) {
            case google.maps.drawing.OverlayType.POLYGON:
                const path = overlay.getPath();
                path.forEach(point => {
                    coordinates.push({ lat: point.lat(), lng: point.lng() });
                });
                break;
                
            case google.maps.drawing.OverlayType.RECTANGLE:
                const bounds = overlay.getBounds();
                const ne = bounds.getNorthEast();
                const sw = bounds.getSouthWest();
                coordinates = [
                    { lat: ne.lat(), lng: ne.lng() },
                    { lat: ne.lat(), lng: sw.lng() },
                    { lat: sw.lat(), lng: sw.lng() },
                    { lat: sw.lat(), lng: ne.lng() }
                ];
                break;
                
            case google.maps.drawing.OverlayType.CIRCLE:
                const center = overlay.getCenter();
                const radius = overlay.getRadius();
                coordinates = {
                    center: { lat: center.lat(), lng: center.lng() },
                    radius: radius
                };
                break;
        }

        // Trigger custom event
        const event_data = new CustomEvent('propertyAreaDrawn', {
            detail: {
                mapId: mapId,
                type: type,
                coordinates: coordinates,
                overlay: overlay
            }
        });
        document.dispatchEvent(event_data);
    }

    // Neighborhood Data
    async loadNeighborhoodData(mapId) {
        try {
            const response = await fetch('/api/neighborhoods/map-data', {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to load neighborhood data');
            }

            const data = await response.json();
            this.addNeighborhoodOverlays(mapId, data);

        } catch (error) {
            console.error('Error loading neighborhood data:', error);
        }
    }

    addNeighborhoodOverlays(mapId, neighborhoodData) {
        const config = this.maps.get(mapId);
        
        neighborhoodData.neighborhoods.forEach(neighborhood => {
            if (neighborhood.boundaries) {
                const polygon = new google.maps.Polygon({
                    paths: neighborhood.boundaries,
                    strokeColor: neighborhood.color || '#3b82f6',
                    strokeOpacity: 0.8,
                    strokeWeight: 2,
                    fillColor: neighborhood.color || '#3b82f6',
                    fillOpacity: 0.2,
                    map: config.instance
                });

                // Add click listener for neighborhood info
                polygon.addListener('click', (e) => {
                    this.showNeighborhoodInfo(mapId, e.latLng, neighborhood);
                });

                this.neighborhoods.set(`${mapId}_${neighborhood.id}`, polygon);
            }
        });
    }

    showNeighborhoodInfo(mapId, position, neighborhood) {
        const config = this.maps.get(mapId);
        const infoWindow = config.infoWindow;
        
        const content = `
            <div class="neighborhood-info">
                <h4>${neighborhood.name}</h4>
                <div class="neighborhood-stats">
                    <div class="stat-item">
                        <label>Avg Home Price:</label>
                        <span>$${neighborhood.avgPrice?.toLocaleString() || 'N/A'}</span>
                    </div>
                    <div class="stat-item">
                        <label>School Rating:</label>
                        <span>${neighborhood.schoolRating || 'N/A'}/10</span>
                    </div>
                    <div class="stat-item">
                        <label>Walk Score:</label>
                        <span>${neighborhood.walkScore || 'N/A'}/100</span>
                    </div>
                </div>
                <div class="neighborhood-amenities">
                    ${neighborhood.amenities ? neighborhood.amenities.slice(0, 3).map(amenity => 
                        `<span class="amenity-tag">${amenity}</span>`
                    ).join('') : ''}
                </div>
            </div>
        `;
        
        infoWindow.setContent(content);
        infoWindow.setPosition(position);
        infoWindow.open(config.instance);
    }

    // Search Functionality
    setupMapSearch(mapId, searchInput) {
        const config = this.maps.get(mapId);
        
        // Setup Google Places Autocomplete
        const autocomplete = new google.maps.places.Autocomplete(searchInput, {
            types: ['address'],
            componentRestrictions: { country: 'us' }
        });

        autocomplete.addListener('place_changed', () => {
            const place = autocomplete.getPlace();
            
            if (!place.geometry) {
                console.error('No geometry found for place:', place.name);
                return;
            }

            // Center map on selected place
            if (place.geometry.viewport) {
                config.instance.fitBounds(place.geometry.viewport);
            } else {
                config.instance.setCenter(place.geometry.location);
                config.instance.setZoom(15);
            }

            // Add marker for searched location
            this.addSearchMarker(mapId, place);
        });
    }

    addSearchMarker(mapId, place) {
        const config = this.maps.get(mapId);
        
        // Remove existing search marker
        if (config.searchMarker) {
            config.searchMarker.setMap(null);
        }

        // Create new search marker
        const marker = new google.maps.Marker({
            position: place.geometry.location,
            map: config.instance,
            title: place.name,
            icon: {
                url: '/static/images/markers/search-location.png',
                scaledSize: new google.maps.Size(32, 32)
            }
        });

        config.searchMarker = marker;

        // Show info about the location
        const infoWindow = new google.maps.InfoWindow({
            content: `
                <div class="search-location-info">
                    <h4>${place.name}</h4>
                    <p>${place.formatted_address}</p>
                    <button class="btn btn-primary btn-sm" 
                            onclick="PropertyHubMaps.searchPropertiesNear(${place.geometry.location.lat()}, ${place.geometry.location.lng()})">
                        Find Properties Nearby
                    </button>
                </div>
            `
        });

        marker.addListener('click', () => {
            infoWindow.open(config.instance, marker);
        });
    }

    // Directions and Routing
    getDirections(destination, origin = null) {
        if (!this.directionsService) return;

        // Get user's current location if no origin provided
        if (!origin && navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(
                (position) => {
                    const userLocation = {
                        lat: position.coords.latitude,
                        lng: position.coords.longitude
                    };
                    this.calculateRoute(userLocation, destination);
                },
                (error) => {
                    console.error('Geolocation error:', error);
                    // Fallback to address input
                    const origin = prompt('Enter your starting address:');
                    if (origin) {
                        this.calculateRoute(origin, destination);
                    }
                }
            );
        } else if (origin) {
            this.calculateRoute(origin, destination);
        }
    }

    calculateRoute(origin, destination) {
        const request = {
            origin: origin,
            destination: destination,
            travelMode: google.maps.TravelMode.DRIVING,
            avoidHighways: false,
            avoidTolls: false
        };

        this.directionsService.route(request, (result, status) => {
            if (status === 'OK') {
                this.showDirectionsModal(result);
            } else {
                console.error('Directions request failed:', status);
                alert('Could not calculate directions. Please try again.');
            }
        });
    }

    showDirectionsModal(directionsResult) {
        const route = directionsResult.routes[0];
        const leg = route.legs[0];
        
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.innerHTML = `
            <div class="modal directions-modal">
                <div class="modal-header">
                    <h3>Directions</h3>
                    <button class="modal-close" onclick="this.closest('.modal-overlay').remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                
                <div class="modal-content">
                    <div class="route-summary">
                        <div class="route-info">
                            <strong>Distance:</strong> ${leg.distance.text}
                            <strong>Duration:</strong> ${leg.duration.text}
                        </div>
                    </div>
                    
                    <div class="directions-map" id="directions-map" style="height: 300px; width: 100%;"></div>
                    
                    <div class="turn-by-turn">
                        <h4>Turn-by-turn directions:</h4>
                        <ol class="directions-list">
                            ${leg.steps.map(step => `
                                <li>${step.instructions}</li>
                            `).join('')}
                        </ol>
                    </div>
                </div>
                
                <div class="modal-actions">
                    <button class="btn btn-primary" onclick="PropertyHubMaps.openInGoogleMaps('${leg.start_address}', '${leg.end_address}')">
                        Open in Google Maps
                    </button>
                    <button class="btn btn-secondary" onclick="this.closest('.modal-overlay').remove()">
                        Close
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(modal);

        // Initialize map in modal
        setTimeout(() => {
            const mapElement = document.getElementById('directions-map');
            if (mapElement) {
                const directionsMap = new google.maps.Map(mapElement, {
                    zoom: 13,
                    center: leg.start_location
                });

                const directionsRenderer = new google.maps.DirectionsRenderer({
                    map: directionsMap
                });

                directionsRenderer.setDirections(directionsResult);
            }
        }, 100);
    }

    // Street View
    showStreetView(lat, lng) {
        const position = { lat: lat, lng: lng };
        
        this.streetViewService.getPanorama({
            location: position,
            radius: 50
        }, (data, status) => {
            if (status === 'OK') {
                this.showStreetViewModal(position, data);
            } else {
                alert('Street View not available for this location');
            }
        });
    }

    showStreetViewModal(position, streetViewData) {
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.innerHTML = `
            <div class="modal street-view-modal">
                <div class="modal-header">
                    <h3>Street View</h3>
                    <button class="modal-close" onclick="this.closest('.modal-overlay').remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                
                <div class="modal-content">
                    <div class="street-view-container" id="street-view" style="height: 400px; width: 100%;"></div>
                </div>
            </div>
        `;

        document.body.appendChild(modal);

        // Initialize Street View
        setTimeout(() => {
            const streetViewElement = document.getElementById('street-view');
            if (streetViewElement) {
                const panorama = new google.maps.StreetViewPanorama(streetViewElement, {
                    position: position,
                    pov: {
                        heading: 34,
                        pitch: 10
                    }
                });
            }
        }, 100);
    }

    // Map Styling
    getMapStyles() {
        return [
            {
                featureType: 'water',
                elementType: 'geometry',
                stylers: [{ color: '#e9e9e9' }, { lightness: 17 }]
            },
            {
                featureType: 'landscape',
                elementType: 'geometry',
                stylers: [{ color: '#f5f5f5' }, { lightness: 20 }]
            },
            {
                featureType: 'road.highway',
                elementType: 'geometry.fill',
                stylers: [{ color: '#ffffff' }, { lightness: 17 }]
            },
            {
                featureType: 'road.highway',
                elementType: 'geometry.stroke',
                stylers: [{ color: '#ffffff' }, { lightness: 29 }, { weight: 0.2 }]
            },
            {
                featureType: 'road.arterial',
                elementType: 'geometry',
                stylers: [{ color: '#ffffff' }, { lightness: 18 }]
            },
            {
                featureType: 'road.local',
                elementType: 'geometry',
                stylers: [{ color: '#ffffff' }, { lightness: 16 }]
            },
            {
                featureType: 'poi',
                elementType: 'geometry',
                stylers: [{ color: '#f5f5f5' }, { lightness: 21 }]
            },
            {
                featureType: 'poi.park',
                elementType: 'geometry',
                stylers: [{ color: '#dedede' }, { lightness: 21 }]
            },
            {
                elementType: 'labels.text.stroke',
                stylers: [{ visibility: 'on' }, { color: '#ffffff' }, { lightness: 16 }]
            },
            {
                elementType: 'labels.text.fill',
                stylers: [{ saturation: 36 }, { color: '#333333' }, { lightness: 40 }]
            },
            {
                elementType: 'labels.icon',
                stylers: [{ visibility: 'off' }]
            },
            {
                featureType: 'transit',
                elementType: 'geometry',
                stylers: [{ color: '#f2f2f2' }, { lightness: 19 }]
            },
            {
                featureType: 'administrative',
                elementType: 'geometry.fill',
                stylers: [{ color: '#fefefe' }, { lightness: 20 }]
            },
            {
                featureType: 'administrative',
                elementType: 'geometry.stroke',
                stylers: [{ color: '#fefefe' }, { lightness: 17 }, { weight: 1.2 }]
            }
        ];
    }

    // Event Listeners
    setupEventListeners() {
        // Custom map events
        document.addEventListener('propertyAreaDrawn', (e) => {
            this.handlePropertyAreaDrawn(e.detail);
        });

        // Window resize handler
        window.addEventListener('resize', this.debounce(() => {
            this.resizeAllMaps();
        }, 250));
    }

    setupMapEventListeners(mapId, config) {
        const map = config.instance;

        // Map bounds changed
        map.addListener('bounds_changed', () => {
            this.handleBoundsChanged(mapId);
        });

        // Map zoom changed
        map.addListener('zoom_changed', () => {
            this.handleZoomChanged(mapId);
        });

        // Map click
        map.addListener('click', (e) => {
            this.handleMapClick(mapId, e);
        });

        // Map drag end
        map.addListener('dragend', () => {
            this.handleMapDragEnd(mapId);
        });
    }

    handleBoundsChanged(mapId) {
        // Optionally load more properties when bounds change
        if (this.shouldLoadMoreProperties(mapId)) {
            this.loadPropertiesInBounds(mapId);
        }
    }

    handleZoomChanged(mapId) {
        const config = this.maps.get(mapId);
        const zoom = config.instance.getZoom();
        
        // Adjust marker visibility or clustering based on zoom level
        if (zoom < 10) {
            this.enableClustering(mapId);
        } else {
            this.disableClustering(mapId);
        }
    }

    handleMapClick(mapId, event) {
        const config = this.maps.get(mapId);
        
        // Close any open info windows
        if (config.infoWindow) {
            config.infoWindow.close();
        }
        
        // Track map interaction
        this.trackMapInteraction('map_clicked', {
            mapId: mapId,
            lat: event.latLng.lat(),
            lng: event.latLng.lng()
        });
    }

    handleMapDragEnd(mapId) {
        // Optionally save map position or load new data
        this.saveMapState(mapId);
    }

    // Utility Methods
    generateMapId() {
        return 'map-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    resizeAllMaps() {
        for (const [mapId, config] of this.maps) {
            google.maps.event.trigger(config.instance, 'resize');
        }
    }

    trackMapInteraction(action, data) {
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackEvent(`map_${action}`, data);
        }
    }

    saveMapState(mapId) {
        const config = this.maps.get(mapId);
        const center = config.instance.getCenter();
        const zoom = config.instance.getZoom();
        
        localStorage.setItem(`map_state_${mapId}`, JSON.stringify({
            center: { lat: center.lat(), lng: center.lng() },
            zoom: zoom
        }));
    }

    loadMapState(mapId) {
        const saved = localStorage.getItem(`map_state_${mapId}`);
        if (saved) {
            return JSON.parse(saved);
        }
        return null;
    }

    // Public API Methods - Called from HTML onclick handlers
    static getDirections(destination) {
        if (window.PropertyHubMaps) {
            window.PropertyHubMaps.getDirections(destination);
        }
    }

    static showStreetView(lat, lng) {
        if (window.PropertyHubMaps) {
            window.PropertyHubMaps.showStreetView(lat, lng);
        }
    }

    static openInGoogleMaps(origin, destination) {
        const url = `https://www.google.com/maps/dir/${encodeURIComponent(origin)}/${encodeURIComponent(destination)}`;
        window.open(url, '_blank');
    }

    static searchPropertiesNear(lat, lng) {
        // Redirect to search page with coordinates
        window.location.href = `/properties/search?lat=${lat}&lng=${lng}&radius=5`;
    }

    // Advanced Features
    async loadPropertiesInBounds(mapId) {
        const config = this.maps.get(mapId);
        const bounds = config.instance.getBounds();
        
        if (!bounds) return;

        const ne = bounds.getNorthEast();
        const sw = bounds.getSouthWest();
        
        try {
            const response = await fetch('/api/properties/in-bounds', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({
                    bounds: {
                        north: ne.lat(),
                        south: sw.lat(),
                        east: ne.lng(),
                        west: sw.lng()
                    }
                })
            });

            if (response.ok) {
                const data = await response.json();
                this.updateMapMarkers(mapId, data.properties);
            }
        } catch (error) {
            console.error('Error loading properties in bounds:', error);
        }
    }

    updateMapMarkers(mapId, properties) {
        const config = this.maps.get(mapId);
        
        // Clear existing markers
        config.markers.forEach(marker => marker.setMap(null));
        config.markers = [];

        // Add new markers
        this.addPropertyMarkers(mapId, properties);
    }

    enableClustering(mapId) {
        const config = this.maps.get(mapId);
        if (config.clusterer && config.markers.length > 0) {
            config.clusterer.addMarkers(config.markers);
        }
    }

    disableClustering(mapId) {
        const config = this.maps.get(mapId);
        if (config.clusterer) {
            config.clusterer.clearMarkers();
        }
    }

    shouldLoadMoreProperties(mapId) {
        // Logic to determine if more properties should be loaded
        return true; // Simplified for this example
    }

    getMapInstance(mapId) {
        const config = this.maps.get(mapId);
        return config ? config.instance : null;
    }

    getAllMaps() {
        return Array.from(this.maps.values());
    }

    destroyMap(mapId) {
        const config = this.maps.get(mapId);
        if (config) {
            // Clean up markers
            config.markers.forEach(marker => marker.setMap(null));
            
            // Clean up info windows
            if (config.infoWindow) {
                config.infoWindow.close();
            }
            
            // Clean up clusterer
            if (config.clusterer) {
                config.clusterer.clearMarkers();
            }
            
            // Remove from maps collection
            this.maps.delete(mapId);
        }
    }
}

// Initialize PropertyHub Maps when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    if (document.querySelector('.property-map')) {
        window.PropertyHubMaps = new PropertyHubMaps();
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubMaps;
}

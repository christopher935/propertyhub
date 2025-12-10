/**
 * PropertyHub Mobile System - FINAL JAVASCRIPT FILE
 * Mobile-optimized functionality, touch interactions, PWA features
 * Responsive design adjustments and mobile-specific real estate tools
 */

class PropertyHubMobile {
    constructor() {
        this.isMobile = this.detectMobileDevice();
        this.isTablet = this.detectTabletDevice();
        this.touchDevice = 'ontouchstart' in window;
        this.orientation = screen.orientation?.type || 'portrait-primary';
        this.viewport = { width: window.innerWidth, height: window.innerHeight };
        this.gestures = new Map();
        this.swipeThreshold = 50;
        this.isOnline = navigator.onLine;
        this.offlineQueue = [];
        
        // PWA configuration
        this.pwaConfig = {
            enableOfflineMode: true,
            enablePushNotifications: true,
            enableInstallPrompt: true,
            enableBackgroundSync: true
        };

        // Mobile navigation state
        this.navigationState = {
            menuOpen: false,
            searchOpen: false,
            filtersOpen: false,
            keyboardOpen: false
        };

        // Touch interaction state
        this.touchState = {
            startX: 0,
            startY: 0,
            startTime: 0,
            isScrolling: false,
            activeGesture: null
        };

        // Mobile-specific features
        this.mobileFeatures = {
            enableHapticFeedback: true,
            enableVoiceSearch: true,
            enableCameraCapture: true,
            enableLocationServices: true,
            enableOfflineCache: true
        };

        this.init();
    }

    async init() {
        Logger.log('📱 PropertyHub Mobile initializing...');
        
        if (!this.isMobileOrTablet()) {
            Logger.log('💻 Desktop detected - mobile features disabled');
            return;
        }

        this.setupViewportMeta();
        this.setupMobileNavigation();
        this.setupTouchGestures();
        this.setupOrientationHandling();
        this.setupKeyboardHandling();
        this.setupOfflineMode();
        this.setupPWAFeatures();
        this.setupMobileOptimizations();
        this.setupDeviceFeatures();
        this.setupEventListeners();
        
        Logger.log('✅ PropertyHub Mobile ready');
    }

    // Device Detection
    detectMobileDevice() {
        const userAgent = navigator.userAgent.toLowerCase();
        const mobileKeywords = [
            'android', 'webos', 'iphone', 'ipad', 'ipod', 'blackberry', 'windows phone'
        ];
        
        return mobileKeywords.some(keyword => userAgent.includes(keyword)) ||
               window.innerWidth <= 768;
    }

    detectTabletDevice() {
        const userAgent = navigator.userAgent.toLowerCase();
        return (userAgent.includes('ipad') || 
                (userAgent.includes('android') && !userAgent.includes('mobile')) ||
                (window.innerWidth > 768 && window.innerWidth <= 1024));
    }

    isMobileOrTablet() {
        return this.isMobile || this.isTablet;
    }

    // Viewport and Layout Management
    setupViewportMeta() {
        let viewportMeta = document.querySelector('meta[name="viewport"]');
        
        if (!viewportMeta) {
            viewportMeta = document.createElement('meta');
            viewportMeta.name = 'viewport';
            document.head.appendChild(viewportMeta);
        }

        viewportMeta.content = 'width=device-width, initial-scale=1.0, user-scalable=yes, maximum-scale=5.0';
    }

    // Mobile Navigation
    setupMobileNavigation() {
        this.createMobileNavigationElements();
        this.setupNavigationInteractions();
    }

    createMobileNavigationElements() {
        // Create mobile menu button if it doesn't exist
        let mobileMenuBtn = document.getElementById('mobile-menu-toggle');
        if (!mobileMenuBtn) {
            mobileMenuBtn = document.createElement('button');
            mobileMenuBtn.id = 'mobile-menu-toggle';
            mobileMenuBtn.className = 'mobile-menu-toggle';
            mobileMenuBtn.innerHTML = `
                <span class="hamburger-line"></span>
                <span class="hamburger-line"></span>
                <span class="hamburger-line"></span>
            `;
            
            const header = document.querySelector('header, .navbar, .main-nav');
            if (header) {
                header.appendChild(mobileMenuBtn);
            }
        }

        // Create mobile search toggle
        let mobileSearchBtn = document.getElementById('mobile-search-toggle');
        if (!mobileSearchBtn && document.querySelector('.search-container')) {
            mobileSearchBtn = document.createElement('button');
            mobileSearchBtn.id = 'mobile-search-toggle';
            mobileSearchBtn.className = 'mobile-search-toggle';
            mobileSearchBtn.innerHTML = '<i class="fas fa-search"></i>';
            
            const header = document.querySelector('header, .navbar, .main-nav');
            if (header) {
                header.appendChild(mobileSearchBtn);
            }
        }

        // Create mobile bottom navigation
        this.createBottomNavigation();
    }

    createBottomNavigation() {
        let bottomNav = document.getElementById('mobile-bottom-nav');
        if (bottomNav) return;

        bottomNav = document.createElement('div');
        bottomNav.id = 'mobile-bottom-nav';
        bottomNav.className = 'mobile-bottom-nav';
        bottomNav.innerHTML = `
            <div class="bottom-nav-item" data-route="/dashboard">
                <i class="fas fa-home"></i>
                <span>Home</span>
            </div>
            <div class="bottom-nav-item" data-route="/properties">
                <i class="fas fa-search"></i>
                <span>Search</span>
            </div>
            <div class="bottom-nav-item" data-route="/bookings">
                <i class="fas fa-calendar"></i>
                <span>Bookings</span>
            </div>
            <div class="bottom-nav-item" data-route="/favorites">
                <i class="fas fa-heart"></i>
                <span>Favorites</span>
            </div>
            <div class="bottom-nav-item" data-route="/profile">
                <i class="fas fa-user"></i>
                <span>Profile</span>
            </div>
        `;

        document.body.appendChild(bottomNav);
        this.updateActiveBottomNavItem();
    }

    setupNavigationInteractions() {
        // Mobile menu toggle
        const menuToggle = document.getElementById('mobile-menu-toggle');
        if (menuToggle) {
            menuToggle.addEventListener('click', () => {
                this.toggleMobileMenu();
            });
        }

        // Mobile search toggle
        const searchToggle = document.getElementById('mobile-search-toggle');
        if (searchToggle) {
            searchToggle.addEventListener('click', () => {
                this.toggleMobileSearch();
            });
        }

        // Bottom navigation
        const bottomNavItems = document.querySelectorAll('.bottom-nav-item');
        bottomNavItems.forEach(item => {
            item.addEventListener('click', () => {
                const route = item.dataset.route;
                if (route) {
                    this.navigateToRoute(route);
                }
            });
        });
    }

    toggleMobileMenu() {
        this.navigationState.menuOpen = !this.navigationState.menuOpen;
        
        const menuToggle = document.getElementById('mobile-menu-toggle');
        const mainNav = document.querySelector('.main-nav, .navbar-nav');
        
        if (menuToggle) {
            menuToggle.classList.toggle('active', this.navigationState.menuOpen);
        }
        
        if (mainNav) {
            mainNav.classList.toggle('mobile-open', this.navigationState.menuOpen);
        }

        // Add body class to prevent scrolling when menu is open
        document.body.classList.toggle('mobile-menu-open', this.navigationState.menuOpen);
        
        // Haptic feedback
        this.triggerHapticFeedback('light');
    }

    toggleMobileSearch() {
        this.navigationState.searchOpen = !this.navigationState.searchOpen;
        
        const searchContainer = document.querySelector('.search-container');
        const searchToggle = document.getElementById('mobile-search-toggle');
        
        if (searchContainer) {
            searchContainer.classList.toggle('mobile-open', this.navigationState.searchOpen);
            
            if (this.navigationState.searchOpen) {
                const searchInput = searchContainer.querySelector('input[type="search"], .search-input');
                if (searchInput) {
                    setTimeout(() => searchInput.focus(), 300);
                }
            }
        }
        
        if (searchToggle) {
            searchToggle.classList.toggle('active', this.navigationState.searchOpen);
        }
    }

    updateActiveBottomNavItem() {
        const currentPath = window.location.pathname;
        const bottomNavItems = document.querySelectorAll('.bottom-nav-item');
        
        bottomNavItems.forEach(item => {
            const route = item.dataset.route;
            const isActive = currentPath.startsWith(route) || 
                            (route === '/dashboard' && currentPath === '/');
            
            item.classList.toggle('active', isActive);
        });
    }

    navigateToRoute(route) {
        window.location.href = route;
        this.triggerHapticFeedback('light');
    }

    // Touch Gestures
    setupTouchGestures() {
        if (!this.touchDevice) return;

        // Property card swipe gestures
        this.setupPropertyCardGestures();
        
        // Image gallery swipe
        this.setupImageGalleryGestures();
        
        // Filter panel gestures
        this.setupFilterPanelGestures();
        
        // Pull-to-refresh
        this.setupPullToRefresh();
    }

    setupPropertyCardGestures() {
        const propertyCards = document.querySelectorAll('.property-card');
        
        propertyCards.forEach(card => {
            this.addSwipeGesture(card, {
                onSwipeLeft: () => this.addToFavorites(card.dataset.propertyId),
                onSwipeRight: () => this.shareProperty(card.dataset.propertyId),
                onDoubleTap: () => this.quickViewProperty(card.dataset.propertyId)
            });
        });
    }

    setupImageGalleryGestures() {
        const galleries = document.querySelectorAll('.image-gallery');
        
        galleries.forEach(gallery => {
            this.addSwipeGesture(gallery, {
                onSwipeLeft: () => this.nextImage(gallery),
                onSwipeRight: () => this.previousImage(gallery),
                onPinch: (scale) => this.zoomImage(gallery, scale)
            });
        });
    }

    setupFilterPanelGestures() {
        const filterPanel = document.querySelector('.filter-panel, .search-filters');
        
        if (filterPanel) {
            this.addSwipeGesture(filterPanel, {
                onSwipeUp: () => this.expandFilters(),
                onSwipeDown: () => this.collapseFilters()
            });
        }
    }

    setupPullToRefresh() {
        let startY = 0;
        let isRefreshing = false;
        const refreshThreshold = 100;
        
        document.addEventListener('touchstart', (e) => {
            if (window.pageYOffset === 0) {
                startY = e.touches[0].pageY;
            }
        });

        document.addEventListener('touchmove', (e) => {
            if (isRefreshing) return;
            
            const currentY = e.touches[0].pageY;
            const pullDistance = currentY - startY;
            
            if (window.pageYOffset === 0 && pullDistance > 0) {
                if (pullDistance > refreshThreshold) {
                    this.showPullToRefreshIndicator();
                }
            }
        });

        document.addEventListener('touchend', (e) => {
            if (isRefreshing) return;
            
            const currentY = e.changedTouches[0].pageY;
            const pullDistance = currentY - startY;
            
            if (window.pageYOffset === 0 && pullDistance > refreshThreshold) {
                this.triggerPullToRefresh();
            } else {
                this.hidePullToRefreshIndicator();
            }
        });
    }

    addSwipeGesture(element, callbacks) {
        let startX, startY, startTime;
        let isScrolling = false;

        element.addEventListener('touchstart', (e) => {
            const touch = e.touches[0];
            startX = touch.pageX;
            startY = touch.pageY;
            startTime = Date.now();
            isScrolling = false;
        });

        element.addEventListener('touchmove', (e) => {
            if (isScrolling) return;
            
            const touch = e.touches[0];
            const diffX = Math.abs(touch.pageX - startX);
            const diffY = Math.abs(touch.pageY - startY);
            
            // Determine if this is a scroll gesture
            if (diffY > diffX && diffY > 10) {
                isScrolling = true;
                return;
            }
            
            // Prevent scrolling for horizontal swipes
            if (diffX > 10) {
                e.preventDefault();
            }
        });

        element.addEventListener('touchend', (e) => {
            if (isScrolling) return;
            
            const touch = e.changedTouches[0];
            const diffX = touch.pageX - startX;
            const diffY = touch.pageY - startY;
            const diffTime = Date.now() - startTime;
            
            // Swipe detection
            if (Math.abs(diffX) > this.swipeThreshold && diffTime < 300) {
                if (diffX > 0 && callbacks.onSwipeRight) {
                    callbacks.onSwipeRight();
                    this.triggerHapticFeedback('light');
                } else if (diffX < 0 && callbacks.onSwipeLeft) {
                    callbacks.onSwipeLeft();
                    this.triggerHapticFeedback('light');
                }
            }
            
            // Vertical swipe detection
            if (Math.abs(diffY) > this.swipeThreshold && diffTime < 300) {
                if (diffY > 0 && callbacks.onSwipeDown) {
                    callbacks.onSwipeDown();
                } else if (diffY < 0 && callbacks.onSwipeUp) {
                    callbacks.onSwipeUp();
                }
            }
            
            // Double tap detection
            if (diffTime < 300 && Math.abs(diffX) < 10 && Math.abs(diffY) < 10) {
                if (this.lastTapTime && (Date.now() - this.lastTapTime) < 300) {
                    if (callbacks.onDoubleTap) {
                        callbacks.onDoubleTap();
                        this.triggerHapticFeedback('medium');
                    }
                }
                this.lastTapTime = Date.now();
            }
        });
    }

    // Orientation Handling
    setupOrientationHandling() {
        window.addEventListener('orientationchange', () => {
            setTimeout(() => {
                this.handleOrientationChange();
            }, 100);
        });

        window.addEventListener('resize', () => {
            this.handleViewportChange();
        });
    }

    handleOrientationChange() {
        this.orientation = screen.orientation?.type || (window.innerWidth > window.innerHeight ? 'landscape' : 'portrait');
        this.viewport = { width: window.innerWidth, height: window.innerHeight };
        
        Logger.log(`📱 Orientation changed to: ${this.orientation}`);
        
        // Adjust layout for orientation
        this.adjustLayoutForOrientation();
        
        // Close mobile menu on orientation change
        if (this.navigationState.menuOpen) {
            this.toggleMobileMenu();
        }

        // Resize maps if present
        if (window.PropertyHubMaps) {
            window.PropertyHubMaps.resizeAllMaps();
        }

        // Update property card layout
        this.updatePropertyCardLayout();
    }

    adjustLayoutForOrientation() {
        const isLandscape = this.orientation.includes('landscape');
        
        document.body.classList.toggle('landscape-mode', isLandscape);
        document.body.classList.toggle('portrait-mode', !isLandscape);

        // Adjust property gallery for landscape mode
        const galleries = document.querySelectorAll('.property-gallery');
        galleries.forEach(gallery => {
            if (isLandscape) {
                gallery.classList.add('landscape-gallery');
            } else {
                gallery.classList.remove('landscape-gallery');
            }
        });
    }

    handleViewportChange() {
        const newViewport = { width: window.innerWidth, height: window.innerHeight };
        const heightDiff = Math.abs(newViewport.height - this.viewport.height);
        
        // Detect keyboard opening/closing (significant height change)
        if (heightDiff > 150) {
            this.navigationState.keyboardOpen = newViewport.height < this.viewport.height;
            this.handleKeyboardToggle();
        }
        
        this.viewport = newViewport;
    }

    handleKeyboardToggle() {
        document.body.classList.toggle('keyboard-open', this.navigationState.keyboardOpen);
        
        if (this.navigationState.keyboardOpen) {
            // Scroll active input into view
            const activeInput = document.activeElement;
            if (activeInput && (activeInput.tagName === 'INPUT' || activeInput.tagName === 'TEXTAREA')) {
                setTimeout(() => {
                    activeInput.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }, 300);
            }
        }
    }

    // Device Features Integration
    setupDeviceFeatures() {
        this.setupLocationServices();
        this.setupCameraCapture();
        this.setupVoiceSearch();
        this.setupContactIntegration();
    }

    setupLocationServices() {
        if (!navigator.geolocation || !this.mobileFeatures.enableLocationServices) return;

        // Auto-detect user location for property search
        const locationButtons = document.querySelectorAll('.use-my-location, .location-detect');
        
        locationButtons.forEach(button => {
            button.addEventListener('click', () => {
                this.getCurrentLocation();
            });
        });
    }

    async getCurrentLocation() {
        try {
            this.showLocationLoading();
            
            const position = await this.getGeolocation();
            const { latitude, longitude } = position.coords;
            
            // Reverse geocode to get address
            const address = await this.reverseGeocode(latitude, longitude);
            
            // Update location fields
            this.updateLocationFields(latitude, longitude, address);
            
            // Update map if present
            if (window.PropertyHubMaps) {
                this.centerMapOnLocation(latitude, longitude);
            }
            
            this.hideLocationLoading();
            this.triggerHapticFeedback('success');
            
        } catch (error) {
            console.error('Location detection failed:', error);
            this.showLocationError(error.message);
            this.hideLocationLoading();
        }
    }

    getGeolocation() {
        return new Promise((resolve, reject) => {
            navigator.geolocation.getCurrentPosition(
                resolve,
                reject,
                {
                    enableHighAccuracy: true,
                    timeout: 10000,
                    maximumAge: 300000 // 5 minutes
                }
            );
        });
    }

    setupCameraCapture() {
        if (!navigator.mediaDevices || !this.mobileFeatures.enableCameraCapture) return;

        const cameraButtons = document.querySelectorAll('.camera-capture, .photo-capture');
        
        cameraButtons.forEach(button => {
            button.addEventListener('click', () => {
                this.openCameraCapture(button.dataset.target);
            });
        });
    }

    async openCameraCapture(targetInput) {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({
                video: { 
                    facingMode: 'environment', // Use back camera
                    width: { ideal: 1920 },
                    height: { ideal: 1080 }
                }
            });

            this.showCameraModal(stream, targetInput);
            
        } catch (error) {
            console.error('Camera access failed:', error);
            
            // Fallback to file input
            const fileInput = document.createElement('input');
            fileInput.type = 'file';
            fileInput.accept = 'image/*';
            fileInput.capture = 'environment';
            fileInput.click();
            
            fileInput.addEventListener('change', (e) => {
                if (e.target.files.length > 0) {
                    this.handleCapturedPhoto(e.target.files[0], targetInput);
                }
            });
        }
    }

    showCameraModal(stream, targetInput) {
        const modal = document.createElement('div');
        modal.className = 'camera-modal-overlay';
        modal.innerHTML = `
            <div class="camera-modal">
                <div class="camera-header">
                    <button class="camera-close" onclick="PropertyHubMobile.closeCameraModal()">
                        <i class="fas fa-times"></i>
                    </button>
                    <h3>Take Photo</h3>
                </div>
                
                <div class="camera-viewport">
                    <video id="camera-video" autoplay playsinline></video>
                </div>
                
                <div class="camera-controls">
                    <button class="camera-capture-btn" onclick="PropertyHubMobile.capturePhoto('${targetInput}')">
                        <i class="fas fa-camera"></i>
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(modal);
        
        const video = document.getElementById('camera-video');
        video.srcObject = stream;
        
        this.currentCameraStream = stream;
    }

    capturePhoto(targetInput) {
        const video = document.getElementById('camera-video');
        const canvas = document.createElement('canvas');
        const context = canvas.getContext('2d');
        
        canvas.width = video.videoWidth;
        canvas.height = video.videoHeight;
        context.drawImage(video, 0, 0, canvas.width, canvas.height);
        
        canvas.toBlob((blob) => {
            const file = new File([blob], `property-photo-${Date.now()}.jpg`, { type: 'image/jpeg' });
            this.handleCapturedPhoto(file, targetInput);
        }, 'image/jpeg', 0.9);
        
        this.closeCameraModal();
        this.triggerHapticFeedback('success');
    }

    closeCameraModal() {
        if (this.currentCameraStream) {
            this.currentCameraStream.getTracks().forEach(track => track.stop());
            this.currentCameraStream = null;
        }
        
        const modal = document.querySelector('.camera-modal-overlay');
        if (modal) {
            modal.remove();
        }
    }

    handleCapturedPhoto(file, targetInput) {
        if (targetInput && window.PropertyHubFileUpload) {
            const uploadZone = document.getElementById(targetInput);
            if (uploadZone) {
                window.PropertyHubFileUpload.handleFileSelection(targetInput, [file]);
            }
        }
    }

    setupVoiceSearch() {
        if (!window.SpeechRecognition && !window.webkitSpeechRecognition) return;
        if (!this.mobileFeatures.enableVoiceSearch) return;

        const voiceButtons = document.querySelectorAll('.voice-search, .mic-search');
        
        voiceButtons.forEach(button => {
            button.addEventListener('click', () => {
                this.startVoiceSearch(button.dataset.target);
            });
        });
    }

    startVoiceSearch(targetInput) {
        const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
        const recognition = new SpeechRecognition();
        
        recognition.continuous = false;
        recognition.interimResults = true;
        recognition.lang = 'en-US';

        const targetElement = document.getElementById(targetInput) || 
                            document.querySelector('.search-input');

        recognition.onstart = () => {
            this.showVoiceSearchIndicator();
            this.triggerHapticFeedback('light');
        };

        recognition.onresult = (event) => {
            const result = event.results[event.results.length - 1];
            if (result.isFinal && targetElement) {
                targetElement.value = result[0].transcript;
                targetElement.dispatchEvent(new Event('input', { bubbles: true }));
            }
        };

        recognition.onend = () => {
            this.hideVoiceSearchIndicator();
        };

        recognition.onerror = (event) => {
            console.error('Voice recognition error:', event.error);
            this.hideVoiceSearchIndicator();
            this.showVoiceSearchError(event.error);
        };

        recognition.start();
    }

    // Offline Mode and PWA
    setupOfflineMode() {
        if (!this.pwaConfig.enableOfflineMode) return;

        // Register service worker
        this.registerServiceWorker();
        
        // Setup offline event listeners
        window.addEventListener('online', () => {
            this.isOnline = true;
            this.handleOnlineStatus();
        });

        window.addEventListener('offline', () => {
            this.isOnline = false;
            this.handleOfflineStatus();
        });

        // Cache critical resources
        this.cacheEssentialResources();
    }

    async registerServiceWorker() {
        if ('serviceWorker' in navigator) {
            try {
                const registration = await navigator.serviceWorker.register('/sw.js');
                Logger.log('✅ Service Worker registered');
                
                // Listen for updates
                registration.addEventListener('updatefound', () => {
                    this.showUpdateAvailableNotification();
                });
                
            } catch (error) {
                console.error('Service Worker registration failed:', error);
            }
        }
    }

    handleOnlineStatus() {
        Logger.log('🌐 Connection restored');
        
        // Process offline queue
        this.processOfflineQueue();
        
        // Show online notification
        this.showConnectionStatus('online');
        
        // Sync data
        this.syncOfflineData();
    }

    handleOfflineStatus() {
        Logger.log('📴 Connection lost');
        
        // Show offline notification
        this.showConnectionStatus('offline');
        
        // Enable offline mode UI
        document.body.classList.add('offline-mode');
    }

    async processOfflineQueue() {
        while (this.offlineQueue.length > 0) {
            const request = this.offlineQueue.shift();
            
            try {
                await fetch(request.url, request.options);
                Logger.log('✅ Offline request processed:', request.url);
            } catch (error) {
                console.error('Failed to process offline request:', error);
                // Re-queue if still offline
                if (!this.isOnline) {
                    this.offlineQueue.unshift(request);
                    break;
                }
            }
        }
    }

    setupPWAFeatures() {
        // Install prompt
        window.addEventListener('beforeinstallprompt', (e) => {
            e.preventDefault();
            this.deferredPrompt = e;
            this.showInstallPrompt();
        });

        // Push notifications
        if ('Notification' in window && this.pwaConfig.enablePushNotifications) {
            this.setupPushNotifications();
        }
    }

    showInstallPrompt() {
        if (!this.pwaConfig.enableInstallPrompt) return;

        const installBanner = document.createElement('div');
        installBanner.className = 'pwa-install-banner';
        installBanner.innerHTML = `
            <div class="install-content">
                <i class="fas fa-mobile-alt"></i>
                <div class="install-text">
                    <h4>Install PropertyHub</h4>
                    <p>Add to your home screen for quick access</p>
                </div>
                <button class="btn btn-primary btn-sm" onclick="PropertyHubMobile.installPWA()">
                    Install
                </button>
                <button class="install-close" onclick="this.parentElement.parentElement.remove()">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;

        document.body.appendChild(installBanner);
    }

    async installPWA() {
        if (this.deferredPrompt) {
            this.deferredPrompt.prompt();
            const { outcome } = await this.deferredPrompt.userChoice;
            
            if (outcome === 'accepted') {
                Logger.log('✅ PWA installed');
                this.trackPWAInstall('accepted');
            } else {
                this.trackPWAInstall('dismissed');
            }
            
            this.deferredPrompt = null;
            
            // Remove install banner
            const banner = document.querySelector('.pwa-install-banner');
            if (banner) banner.remove();
        }
    }

    // Mobile Optimizations
    setupMobileOptimizations() {
        // Optimize touch targets
        this.optimizeTouchTargets();
        
        // Setup lazy loading for images
        this.setupLazyLoading();
        
        // Optimize form inputs
        this.optimizeFormInputs();
        
        // Setup infinite scroll
        this.setupInfiniteScroll();
    }

    optimizeTouchTargets() {
        const smallButtons = document.querySelectorAll('button, .btn, .clickable');
        
        smallButtons.forEach(button => {
            const rect = button.getBoundingClientRect();
            
            if (rect.width < 44 || rect.height < 44) {
                button.classList.add('touch-target-small');
                button.style.minWidth = '44px';
                button.style.minHeight = '44px';
            }
        });
    }

    setupLazyLoading() {
        if ('IntersectionObserver' in window) {
            const imageObserver = new IntersectionObserver((entries, observer) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const img = entry.target;
                        if (img.dataset.src) {
                            img.src = img.dataset.src;
                            img.classList.remove('lazy');
                            observer.unobserve(img);
                        }
                    }
                });
            });

            document.querySelectorAll('img[data-src]').forEach(img => {
                imageObserver.observe(img);
            });
        }
    }

    optimizeFormInputs() {
        const inputs = document.querySelectorAll('input, textarea');
        
        inputs.forEach(input => {
            // Add appropriate input types and attributes for mobile
            if (input.name === 'email' || input.type === 'email') {
                input.autocomplete = 'email';
                input.autocapitalize = 'none';
                input.autocorrect = 'off';
            }
            
            if (input.name === 'phone' || input.classList.contains('phone-input')) {
                input.type = 'tel';
                input.autocomplete = 'tel';
            }
            
            if (input.name === 'zipCode' || input.classList.contains('zip-input')) {
                input.type = 'text';
                input.pattern = '[0-9]{5}(-[0-9]{4})?';
                input.autocomplete = 'postal-code';
            }
            
            if (input.classList.contains('currency-input')) {
                input.type = 'text';
                input.inputMode = 'decimal';
            }
        });
    }

    setupInfiniteScroll() {
        const scrollContainers = document.querySelectorAll('.infinite-scroll, .property-results');
        
        scrollContainers.forEach(container => {
            const observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        this.loadMoreContent(container);
                    }
                });
            }, { threshold: 0.1 });

            // Create sentinel element
            const sentinel = document.createElement('div');
            sentinel.className = 'scroll-sentinel';
            container.appendChild(sentinel);
            
            observer.observe(sentinel);
        });
    }

    async loadMoreContent(container) {
        if (container.dataset.loading === 'true') return;
        
        container.dataset.loading = 'true';
        
        try {
            const page = parseInt(container.dataset.page || '1') + 1;
            const response = await fetch(`${container.dataset.loadUrl}?page=${page}`, {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.appendContent(container, data.content);
                container.dataset.page = page.toString();
            }
        } catch (error) {
            console.error('Failed to load more content:', error);
        } finally {
            container.dataset.loading = 'false';
        }
    }

    // Haptic Feedback
    triggerHapticFeedback(type = 'light') {
        if (!this.mobileFeatures.enableHapticFeedback) return;
        
        if (navigator.vibrate) {
            const patterns = {
                light: [10],
                medium: [20],
                heavy: [50],
                success: [10, 10, 10],
                error: [100, 50, 100]
            };
            
            navigator.vibrate(patterns[type] || patterns.light);
        }
    }

    // Utility Methods
    showLocationLoading() {
        const buttons = document.querySelectorAll('.use-my-location, .location-detect');
        buttons.forEach(button => {
            button.disabled = true;
            button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Getting location...';
        });
    }

    hideLocationLoading() {
        const buttons = document.querySelectorAll('.use-my-location, .location-detect');
        buttons.forEach(button => {
            button.disabled = false;
            button.innerHTML = '<i class="fas fa-map-marker-alt"></i> Use My Location';
        });
    }

    showLocationError(message) {
        if (window.PropertyHubNotifications) {
            window.PropertyHubNotifications.showToastNotification({
                title: 'Location Error',
                message: message,
                type: 'error',
                duration: 5000
            });
        }
    }

    showConnectionStatus(status) {
        const statusElement = document.getElementById('connection-status') || this.createConnectionStatus();
        
        statusElement.className = `connection-status ${status}`;
        statusElement.textContent = status === 'online' ? 'Connected' : 'Offline';
        statusElement.style.display = 'block';
        
        // Auto-hide online status after 3 seconds
        if (status === 'online') {
            setTimeout(() => {
                statusElement.style.display = 'none';
            }, 3000);
        }
    }

    createConnectionStatus() {
        const statusElement = document.createElement('div');
        statusElement.id = 'connection-status';
        statusElement.className = 'connection-status';
        document.body.appendChild(statusElement);
        return statusElement;
    }

    updatePropertyCardLayout() {
        const propertyCards = document.querySelectorAll('.property-card');
        const isLandscape = this.orientation.includes('landscape');
        
        propertyCards.forEach(card => {
            if (isLandscape) {
                card.classList.add('landscape-card');
            } else {
                card.classList.remove('landscape-card');
            }
        });
    }

    trackPWAInstall(outcome) {
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackEvent('pwa_install', {
                outcome: outcome,
                platform: navigator.platform,
                userAgent: navigator.userAgent
            });
        }
    }

    setupEventListeners() {
        // Page visibility for mobile app lifecycle
        document.addEventListener('visibilitychange', () => {
            if (document.visibilityState === 'visible') {
                this.handleAppResume();
            } else {
                this.handleAppPause();
            }
        });

        // Handle app installation
        window.addEventListener('appinstalled', () => {
            Logger.log('✅ PropertyHub PWA installed');
            this.trackPWAInstall('installed');
        });
    }

    handleAppResume() {
        // Refresh data when app becomes visible
        if (window.PropertyHubDashboard && window.PropertyHubDashboard.refreshDashboard) {
            window.PropertyHubDashboard.refreshDashboard();
        }
    }

    handleAppPause() {
        // Save any pending data when app becomes hidden
        this.savePendingData();
    }

    savePendingData() {
        // Save any form data or state that should persist
        const forms = document.querySelectorAll('form[data-autosave="true"]');
        forms.forEach(form => {
            const formData = new FormData(form);
            const data = {};
            for (const [key, value] of formData) {
                data[key] = value;
            }
            localStorage.setItem(`autosave_${form.id}`, JSON.stringify(data));
        });
    }

    // Public API Methods - Called from HTML onclick handlers
    static toggleMobileMenu() {
        if (window.PropertyHubMobile) {
            window.PropertyHubMobile.toggleMobileMenu();
        }
    }

    static closeCameraModal() {
        if (window.PropertyHubMobile) {
            window.PropertyHubMobile.closeCameraModal();
        }
    }

    static capturePhoto(targetInput) {
        if (window.PropertyHubMobile) {
            window.PropertyHubMobile.capturePhoto(targetInput);
        }
    }

    static installPWA() {
        if (window.PropertyHubMobile) {
            window.PropertyHubMobile.installPWA();
        }
    }

    static getCurrentLocation() {
        if (window.PropertyHubMobile) {
            window.PropertyHubMobile.getCurrentLocation();
        }
    }

    static startVoiceSearch(targetInput) {
        if (window.PropertyHubMobile) {
            window.PropertyHubMobile.startVoiceSearch(targetInput);
        }
    }

    // Mobile-specific property features
    addToFavorites(propertyId) {
        if (window.PropertyHubSearch) {
            window.PropertyHubSearch.toggleFavorite(propertyId);
        }
        this.triggerHapticFeedback('light');
    }

    shareProperty(propertyId) {
        if (navigator.share) {
            const property = this.getPropertyData(propertyId);
            navigator.share({
                title: `Property: ${property.address}`,
                text: `Check out this property on PropertyHub`,
                url: `${window.location.origin}/properties/${propertyId}`
            });
        } else {
            // Fallback to clipboard
            const url = `${window.location.origin}/properties/${propertyId}`;
            navigator.clipboard.writeText(url).then(() => {
                this.showToast('Property link copied to clipboard');
            });
        }
        this.triggerHapticFeedback('success');
    }

    quickViewProperty(propertyId) {
        // Open property quick view modal optimized for mobile
        this.showPropertyQuickView(propertyId);
        this.triggerHapticFeedback('medium');
    }

    showToast(message) {
        if (window.PropertyHubNotifications) {
            window.PropertyHubNotifications.showToastNotification({
                title: 'PropertyHub',
                message: message,
                type: 'info',
                duration: 3000
            });
        }
    }

    // Performance optimizations
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

    throttle(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }
}

// Initialize PropertyHub Mobile when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.PropertyHubMobile = new PropertyHubMobile();
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubMobile;
}

function searchJourney() {
    return {
        viewedProperties: parseInt(localStorage.getItem('viewed_count') || '0'),
        savedProperties: JSON.parse(localStorage.getItem('saved_properties') || '[]').length,
        hasBooked: localStorage.getItem('has_booking') === 'true',
        
        get hasActivity() {
            return this.viewedProperties > 0;
        },
        
        getEncouragement() {
            if (this.savedProperties === 0) return "Save properties you like to compare them later!";
            if (this.savedProperties > 0 && !this.hasBooked) return `You have ${this.savedProperties} saved - ready to schedule a tour?`;
            return "You're on your way to finding your perfect home!";
        }
    };
}

function continueBanner() {
    return {
        get hasIncompleteAction() {
            if (localStorage.getItem('incomplete_booking')) return true;
            try {
                const saved = JSON.parse(localStorage.getItem('saved_properties') || '[]');
                return saved.length > 0;
            } catch {
                return false;
            }
        },
        get bannerTitle() {
            if (localStorage.getItem('incomplete_booking')) return "You have an unfinished booking";
            if (localStorage.getItem('saved_properties')) return "You have saved properties waiting";
            return "";
        },
        get bannerSubtitle() {
            if (localStorage.getItem('incomplete_booking')) return "Complete your booking to secure your tour";
            if (localStorage.getItem('saved_properties')) {
                const saved = JSON.parse(localStorage.getItem('saved_properties') || '[]');
                return `${saved.length} properties saved for you`;
            }
            return "";
        },
        get continueUrl() {
            if (localStorage.getItem('incomplete_booking')) return "/book-showing";
            if (localStorage.getItem('saved_properties')) return "/properties";
            return "/";
        },
        dismiss() {
            localStorage.removeItem('incomplete_booking');
            this.hasIncompleteAction = false;
        }
    };
}

function attentionBadge() {
    return {
        needsAttention: false,
        attentionCount: 0,
        
        async init() {
            await this.checkAttention();
        },
        
        async checkAttention() {
            const hotLeads = parseInt(document.querySelector('[data-hot-leads]')?.dataset.hotLeads || '0');
            const pendingBookings = parseInt(document.querySelector('[data-pending-bookings]')?.dataset.pendingBookings || '0');
            
            this.attentionCount = hotLeads + pendingBookings;
            this.needsAttention = this.attentionCount > 0;
        }
    };
}

function bookingCelebration() {
    return {
        showCelebration: false,
        celebrationMessage: "",
        dealValue: "$0",
        daysToClose: "0",
        
        celebrate() {
            const confettiContainer = document.querySelector('.confetti-container');
            if (confettiContainer) {
                for (let i = 0; i < 50; i++) {
                    const confetti = document.createElement('div');
                    confetti.className = 'confetti';
                    confetti.style.left = `${Math.random() * 100}%`;
                    confetti.style.animationDelay = `${Math.random() * 3}s`;
                    confetti.style.background = ['#C5A572', '#1E3A5F', '#10b981'][Math.floor(Math.random() * 3)];
                    confettiContainer.appendChild(confetti);
                }
                
                setTimeout(() => {
                    this.showCelebration = false;
                }, 5000);
            }
        },
        
        shareOnTwitter() {
            const text = encodeURIComponent("I just scheduled a tour for my next home! 🏠");
            window.open(`https://twitter.com/intent/tweet?text=${text}`, '_blank');
        },
        
        copyLink() {
            navigator.clipboard.writeText(window.location.href);
            alert('Link copied to clipboard!');
        }
    };
}

function isNew(createdAt) {
    const daysSinceListing = (Date.now() - new Date(createdAt)) / (1000 * 60 * 60 * 24);
    return daysSinceListing < 7;
}

function trackPropertyView(propertyId) {
    let viewCount = parseInt(localStorage.getItem('viewed_count') || '0');
    viewCount++;
    localStorage.setItem('viewed_count', viewCount.toString());
    
    let viewedProperties = JSON.parse(localStorage.getItem('viewed_properties') || '[]');
    if (!viewedProperties.includes(propertyId)) {
        viewedProperties.push(propertyId);
        localStorage.setItem('viewed_properties', JSON.stringify(viewedProperties));
    }
}

function adminProgressWidget() {
    return {
        completionPercent: 0,
        tasks: {
            leadsContacted: false,
            bookingsConfirmed: false,
            followUpsCompleted: false
        },
        
        async init() {
            await this.calculateProgress();
        },
        
        async calculateProgress() {
            const completed = Object.values(this.tasks).filter(Boolean).length;
            const total = Object.keys(this.tasks).length;
            this.completionPercent = Math.round((completed / total) * 100);
        }
    };
}

function dealClosedCelebration() {
    return {
        showCelebration: false,
        celebrationMessage: "",
        dealValue: "$0",
        daysToClose: "0",
        
        triggerCelebration(message, value, days) {
            this.celebrationMessage = message;
            this.dealValue = value;
            this.daysToClose = days;
            this.showCelebration = true;
        }
    };
}

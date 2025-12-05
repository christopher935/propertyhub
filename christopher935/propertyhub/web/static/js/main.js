// PropertyHub Main JavaScript

// Safe mode integration
window.PropertyHub = {
    safeMode: window.SAFE_MODE || false,
    testingMode: window.TESTING_MODE || false,
    mode: window.MODE || 'PRODUCTION',
    
    init: function() {
        console.log(`PropertyHub initialized - Mode: ${this.mode}`);
        if (this.safeMode) {
            console.log('🛡️ Safe Mode: Customer notifications disabled');
        }
        this.setupFormHandlers();
    },
    
    setupFormHandlers: function() {
        // Add form submission handlers
        const forms = document.querySelectorAll('form[action*="/book-showing"], form[action*="/quick-contact"]');
        forms.forEach(form => {
            form.addEventListener('submit', this.handleFormSubmit.bind(this));
        });
    },
    
    handleFormSubmit: function(event) {
        if (this.safeMode) {
            console.log('🧪 Safe Mode: Form submitted (testing mode)');
        }
        
        // Add any additional form validation here
        const form = event.target;
        const submitButton = form.querySelector('button[type="submit"]');
        
        if (submitButton) {
            const originalText = submitButton.textContent;
            submitButton.textContent = 'Processing...';
            submitButton.disabled = true;
            
            // Re-enable button after a delay (in case of errors)
            setTimeout(() => {
                submitButton.textContent = originalText;
                submitButton.disabled = false;
            }, 5000);
        }
    }
};

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    window.PropertyHub.init();
});

// Agent fields toggle for booking form
function toggleAgentFields() {
    const hasAgent = document.getElementById('hasAgent');
    const agentFields = document.getElementById('agentFields');
    
    if (hasAgent && agentFields) {
        const isVisible = hasAgent.value === 'yes';
        agentFields.style.display = isVisible ? 'block' : 'none';
        
        // Set required attributes for agent fields
        const agentInputs = agentFields.querySelectorAll('input');
        agentInputs.forEach(input => {
            input.required = isVisible;
        });
    }
}

// Quick contact modal functions
function openQuickContact() {
    const modal = document.getElementById('quickContactModal');
    if (modal) {
        modal.style.display = 'block';
    }
}

function closeQuickContact() {
    const modal = document.getElementById('quickContactModal');
    if (modal) {
        modal.style.display = 'none';
    }
}

// Close modal when clicking outside of it
window.onclick = function(event) {
    const modal = document.getElementById('quickContactModal');
    if (event.target === modal) {
        closeQuickContact();
    }
}

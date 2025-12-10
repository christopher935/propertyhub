document.addEventListener('alpine:init', () => {
    const srAnnounce = (message, priority = 'polite') => {
        const announcer = document.getElementById('sr-announcer');
        if (announcer) {
            announcer.setAttribute('aria-live', priority);
            announcer.textContent = message;
            setTimeout(() => {
                announcer.textContent = '';
            }, 1000);
        }
    };

    const originalBookingForm = window.bookingForm;
    window.bookingForm = function() {
        const form = originalBookingForm();
        
        const originalNextStep = form.nextStep;
        form.nextStep = function() {
            if (this.step === 1 && this.allRequirementsMet) {
                const button = document.getElementById('step1-next');
                if (button) {
                    button.classList.add('success-animation');
                    setTimeout(() => {
                        button.classList.remove('success-animation');
                    }, 600);
                }
                
                srAnnounce('Pre-qualification complete! Moving to date and time selection.', 'assertive');
                
                setTimeout(() => {
                    originalNextStep.call(this);
                }, 300);
            } else {
                originalNextStep.call(this);
            }
        };
        
        return new Proxy(form, {
            get(target, prop) {
                if (prop === 'requirementsChecked') {
                    return new Proxy(target.requirementsChecked, {
                        set(reqTarget, reqProp, value) {
                            const oldValue = reqTarget[reqProp];
                            reqTarget[reqProp] = value;
                            
                            if (oldValue !== value) {
                                const count = Object.values(reqTarget).filter(v => v).length;
                                
                                const requirementNames = {
                                    income: 'Income requirement',
                                    felony: 'Criminal background check',
                                    eviction: 'Rental history',
                                    credit: 'Credit score'
                                };
                                
                                const message = value 
                                    ? `${requirementNames[reqProp]} confirmed. ${count} of 4 requirements checked.`
                                    : `${requirementNames[reqProp]} unchecked. ${count} of 4 requirements checked.`;
                                
                                srAnnounce(message);
                                
                                if (count === 4) {
                                    setTimeout(() => {
                                        srAnnounce("All requirements confirmed! You're ready to proceed to the next step.", 'assertive');
                                    }, 500);
                                }
                            }
                            
                            return true;
                        }
                    });
                }
                return target[prop];
            }
        });
    };
});

if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', () => {
        const body = document.body;
        if (!document.getElementById('sr-announcer')) {
            const announcer = document.createElement('div');
            announcer.id = 'sr-announcer';
            announcer.setAttribute('role', 'status');
            announcer.setAttribute('aria-live', 'polite');
            announcer.setAttribute('aria-atomic', 'true');
            announcer.style.cssText = 'position: absolute; left: -10000px; width: 1px; height: 1px; overflow: hidden;';
            body.appendChild(announcer);
        }
    });
}

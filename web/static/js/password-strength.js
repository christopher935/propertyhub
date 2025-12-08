function passwordStrength() {
    return {
        password: '',
        strengthPercent: 0,
        strengthClass: '',
        strengthLabel: '',
        hasLength: false,
        hasLower: false,
        hasUpper: false,
        hasNumber: false,
        hasSpecial: false,
        
        init() {
            this.checkStrength();
        },
        
        checkStrength() {
            const pwd = this.password;
            
            this.hasLength = pwd.length >= 8;
            this.hasLower = /[a-z]/.test(pwd);
            this.hasUpper = /[A-Z]/.test(pwd);
            this.hasNumber = /\d/.test(pwd);
            this.hasSpecial = /[!@#$%^&*(),.?":{}|<>]/.test(pwd);
            
            const criteria = [
                this.hasLower,
                this.hasUpper,
                this.hasNumber,
                this.hasSpecial
            ].filter(Boolean).length;
            
            if (!this.hasLength) {
                this.strengthPercent = Math.min(pwd.length / 8 * 20, 20);
                this.strengthClass = 'weak';
                this.strengthLabel = 'Weak';
            } else if (criteria === 1) {
                this.strengthPercent = 30;
                this.strengthClass = 'fair';
                this.strengthLabel = 'Fair';
            } else if (criteria === 2) {
                this.strengthPercent = 50;
                this.strengthClass = 'good';
                this.strengthLabel = 'Good';
            } else if (criteria === 3) {
                this.strengthPercent = 70;
                this.strengthClass = 'strong';
                this.strengthLabel = 'Strong';
            } else if (criteria === 4 && this.hasLength) {
                this.strengthPercent = 100;
                this.strengthClass = 'very-strong';
                this.strengthLabel = 'Very Strong';
            }
            
            if (pwd.length === 0) {
                this.strengthPercent = 0;
                this.strengthClass = '';
                this.strengthLabel = '';
            }
        }
    };
}

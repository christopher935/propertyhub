function mfaPromptData() {
    return {
        dismissed: false,
        hasMFA: false,
        dismissCount: 0,
        
        init() {
            const dismissData = localStorage.getItem('mfaPromptDismissed');
            if (dismissData) {
                const data = JSON.parse(dismissData);
                this.dismissCount = data.count || 0;
                
                if (this.dismissCount >= 3) {
                    this.dismissed = true;
                    return;
                }
                
                if (data.remindAfter) {
                    const remindDate = new Date(data.remindAfter);
                    if (remindDate > new Date()) {
                        this.dismissed = true;
                    }
                }
            }
            
            this.checkMFAStatus();
        },
        
        checkMFAStatus() {
            fetch('/api/admin/mfa/status', {
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                this.hasMFA = data.enabled || false;
            })
            .catch(err => {
                console.error('Failed to check MFA status:', err);
            });
        },
        
        remindLater() {
            const remindAfter = new Date();
            remindAfter.setDate(remindAfter.getDate() + 7);
            
            const dismissCount = this.dismissCount + 1;
            
            localStorage.setItem('mfaPromptDismissed', JSON.stringify({
                count: dismissCount,
                remindAfter: remindAfter.toISOString()
            }));
            
            this.dismissed = true;
            this.dismissCount = dismissCount;
        },
        
        dismissMFA() {
            const dismissCount = this.dismissCount + 1;
            
            localStorage.setItem('mfaPromptDismissed', JSON.stringify({
                count: dismissCount
            }));
            
            this.dismissed = true;
            this.dismissCount = dismissCount;
        }
    };
}

function mfaStatusData() {
    return {
        mfaEnabled: false,
        backupCodesRemaining: 0,
        loading: true,
        
        init() {
            this.fetchMFAStatus();
        },
        
        fetchMFAStatus() {
            fetch('/api/admin/mfa/status', {
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                this.mfaEnabled = data.enabled || false;
                this.backupCodesRemaining = data.backupCodesRemaining || 0;
                this.loading = false;
            })
            .catch(err => {
                console.error('Failed to fetch MFA status:', err);
                this.loading = false;
            });
        }
    };
}

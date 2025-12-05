/**
 * PropertyHub Dashboard - Interactive Behavior System
 * Implements: AI suggestion rotation, card expansion, button states, toast notifications
 * Methodology: Cosmetic Polish + Behavioral Excellence = Pixel-Perfect Excellence
 */

// ============================================
// DASHBOARD STATE MANAGEMENT
// ============================================

const dashboardState = {
  suggestions: {
    queue: [
      {
        id: 1,
        message: 'Good afternoon, Chris! You have <strong>31 hot leads</strong> right now - that\'s <strong>15% higher</strong> than last week. I noticed 8 of them viewed luxury properties over $3,000/month. Want me to send them the new downtown listings?',
        actions: [
          { text: 'Yes, Send Campaign', type: 'primary', action: 'sendCampaign' },
          { text: 'Show Me Details', type: 'secondary', action: 'showDetails' },
          { text: 'Not Now', type: 'secondary', action: 'dismiss' }
        ]
      },
      {
        id: 2,
        message: 'Lead <strong>Mike Chen</strong> is <strong>78% likely</strong> to apply within 3 days if you show him pet-friendly properties under $2,500. I found 4 matches.',
        actions: [
          { text: 'Send Listings to Mike', type: 'primary', action: 'sendListings' },
          { text: 'View Profile', type: 'secondary', action: 'viewProfile' },
          { text: 'Not Now', type: 'secondary', action: 'dismiss' }
        ]
      },
      {
        id: 3,
        message: '<strong>12 leads</strong> went cold this week. I\'ve prepared a re-engagement campaign called "New Properties Alert" with your latest listings. Ready to send?',
        actions: [
          { text: 'Yes, Start Campaign', type: 'primary', action: 'startCampaign' },
          { text: 'Review Leads', type: 'secondary', action: 'reviewLeads' },
          { text: 'Not Now', type: 'secondary', action: 'dismiss' }
        ]
      }
    ],
    currentIndex: 0
  }
};

// ============================================
// TOAST NOTIFICATION SYSTEM
// ============================================

function showToast(message, type = 'success', duration = 3000) {
  const container = document.getElementById('toast-container') || createToastContainer();
  
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = `
    <div class="toast-icon">${type === 'success' ? '✓' : type === 'error' ? '✕' : 'ℹ'}</div>
    <div class="toast-message">${message}</div>
    <button class="toast-close" onclick="this.parentElement.remove()">×</button>
  `;
  
  container.appendChild(toast);
  
  setTimeout(() => {
    toast.style.animation = 'slideOutRight 300ms ease-out';
    setTimeout(() => toast.remove(), 300);
  }, duration);
}

function createToastContainer() {
  const container = document.createElement('div');
  container.id = 'toast-container';
  container.className = 'toast-container';
  document.body.appendChild(container);
  return container;
}

// ============================================
// BUTTON LOADING STATE UTILITY
// ============================================

function setButtonLoading(button, loading = true) {
  if (loading) {
    button.disabled = true;
    button.classList.add('loading');
    button.dataset.originalText = button.innerHTML;
    button.innerHTML = '<span class="btn-spinner"></span>';
  } else {
    button.disabled = false;
    button.classList.remove('loading');
    button.innerHTML = button.dataset.originalText || button.innerHTML;
  }
}

function setButtonSuccess(button, text = '✓ Done') {
  button.classList.remove('loading');
  button.classList.add('success');
  button.innerHTML = text;
  button.disabled = true;
}

// ============================================
// AI SUGGESTION ROTATION SYSTEM
// ============================================

function initAISuggestions() {
  const aiNarration = document.querySelector('.ai-narration');
  if (!aiNarration) return;
  
  const actionButtons = aiNarration.querySelectorAll('.btn');
  actionButtons.forEach(button => {
    button.addEventListener('click', (e) => handleSuggestionAction(e, button));
  });
}

async function handleSuggestionAction(event, button) {
  event.preventDefault();
  
  setButtonLoading(button);
  
  try {
    await sleep(800);
    setButtonSuccess(button);
    showToast('Action completed successfully!', 'success');
    await sleep(1500);
    rotateSuggestion();
  } catch (error) {
    showToast('Action failed. Please try again.', 'error');
  }
}

function rotateSuggestion() {
  const aiNarration = document.querySelector('.ai-narration');
  if (!aiNarration) return;
  
  aiNarration.classList.add('fade-out');
  
  setTimeout(() => {
    dashboardState.suggestions.currentIndex = 
      (dashboardState.suggestions.currentIndex + 1) % dashboardState.suggestions.queue.length;
    
    const nextSuggestion = dashboardState.suggestions.queue[dashboardState.suggestions.currentIndex];
    updateSuggestionContent(nextSuggestion);
    
    aiNarration.classList.remove('fade-out');
    aiNarration.classList.add('fade-in');
    
    setTimeout(() => aiNarration.classList.remove('fade-in'), 300);
  }, 300);
}

function updateSuggestionContent(suggestion) {
  const messageEl = document.querySelector('.ai-message');
  const actionsEl = document.querySelector('.ai-actions');
  
  if (messageEl) messageEl.innerHTML = suggestion.message;
  
  if (actionsEl) {
    actionsEl.innerHTML = suggestion.actions.map(action => `
      <button class="btn ${action.type === 'secondary' ? 'btn-secondary' : ''}">${action.text}</button>
    `).join('');
    
    actionsEl.querySelectorAll('.btn').forEach(button => {
      button.addEventListener('click', (e) => handleSuggestionAction(e, button));
    });
  }
}

// ============================================
// QUICK STATS CARD EXPANSION
// ============================================

function initQuickStatsCards() {
  const statCards = document.querySelectorAll('.stat-card');
  
  statCards.forEach(card => {
    const buttons = card.querySelectorAll('.btn');
    buttons.forEach(button => {
      button.addEventListener('click', (e) => handleCardAction(e, card, button));
    });
  });
}

async function handleCardAction(event, card, button) {
  event.preventDefault();
  
  setButtonLoading(button);
  
  try {
    await sleep(1000);
    setButtonSuccess(button, '✓ Done');
    showToast('Action completed successfully!', 'success');
  } catch (error) {
    showToast('Action failed. Please try again.', 'error');
  }
}

// ============================================
// PREDICTIVE INSIGHTS INTERACTIONS
// ============================================

function initPredictiveInsights() {
  const insightCards = document.querySelectorAll('.insight-card');
  
  insightCards.forEach(card => {
    const buttons = card.querySelectorAll('.btn');
    buttons.forEach(button => {
      button.addEventListener('click', (e) => handleInsightAction(e, card, button));
    });
  });
}

async function handleInsightAction(event, card, button) {
  event.preventDefault();
  
  const isPrimary = !button.classList.contains('btn-secondary');
  
  setButtonLoading(button);
  
  try {
    await sleep(1000);
    
    if (isPrimary) {
      setButtonSuccess(button);
      showToast('Action completed successfully!', 'success');
      
      await sleep(1500);
      
      card.style.opacity = '0';
      card.style.transform = 'translateY(-20px)';
      card.style.transition = 'all 400ms ease-out';
      
      setTimeout(() => card.remove(), 400);
    } else {
      setButtonSuccess(button);
      showToast('Loading details...', 'info');
    }
  } catch (error) {
    showToast('Action failed. Please try again.', 'error');
  }
}

// ============================================
// UTILITY FUNCTIONS
// ============================================

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// ============================================
// INITIALIZATION
// ============================================

document.addEventListener('DOMContentLoaded', () => {
  
  initAISuggestions();
  initQuickStatsCards();
  initPredictiveInsights();
  
});

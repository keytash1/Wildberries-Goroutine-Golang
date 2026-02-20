let startTime;

async function getOrder() {
    const orderId = document.getElementById('orderId').value.trim();
    
    if (!orderId) {
        showError('Please enter an order ID');
        return;
    }

    startTime = performance.now();
    
    showLoading();
    hideError();
    hideCacheInfo();
    hideOrderDetails();

    try {
        const response = await fetch(`/order/${orderId}`);
        const data = await response.json();
        const endTime = performance.now();
        const responseTime = endTime - startTime;
        
        if (!response.ok) {
            throw new Error(data.error || 'Error fetching order');
        }
        
        displayOrder(data, responseTime);
    } catch (error) {
        showError(error.message);
    } finally {
        hideLoading();
    }
}

function displayOrder(order, responseTime) {
    const fromCache = responseTime < 20; 
    
    showCacheInfo(fromCache, responseTime);
    
    const detailsDiv = document.getElementById('orderDetails');
    detailsDiv.innerHTML = `<pre>${JSON.stringify(order, null, 2)}</pre>`;
    detailsDiv.classList.remove('hidden');
}

function showCacheInfo(fromCache, responseTime) {
    const infoDiv = document.getElementById('cacheInfo');
    infoDiv.className = `cache-info ${fromCache ? 'hit' : 'miss'}`;
    infoDiv.innerHTML = fromCache 
        ? ` CACHE HIT (${responseTime.toFixed(2)} ms)` 
        : ` CACHE MISS (${responseTime.toFixed(2)} ms) - loaded from database`;
    infoDiv.classList.remove('hidden');
}

function showLoading() {
    document.getElementById('loading').classList.remove('hidden');
}

function hideLoading() {
    document.getElementById('loading').classList.add('hidden');
}

function showError(message) {
    const errorDiv = document.getElementById('error');
    errorDiv.textContent = message;
    errorDiv.classList.remove('hidden');
}

function hideError() {
    document.getElementById('error').classList.add('hidden');
}

function hideCacheInfo() {
    document.getElementById('cacheInfo').classList.add('hidden');
}

function hideOrderDetails() {
    document.getElementById('orderDetails').classList.add('hidden');
}

// Загружаем при старте
document.addEventListener('DOMContentLoaded', getOrder);

// Enter в поле ввода
document.getElementById('orderId').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') getOrder();
});
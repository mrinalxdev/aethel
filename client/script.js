async function fetchAnalytics() {
    const loading = document.getElementById('loading');
    const tableBody = document.getElementById('analytics-body');
    loading.classList.remove('hidden');
    tableBody.innerHTML = '';

    try {
        const response = await fetch('http://localhost:8080/analytics/users');
        const data = await response.json();
        data.forEach(item => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td class="p-2 border">${item.user_id}</td>
                <td class="p-2 border">${item.event_count}</td>
            `;
            tableBody.appendChild(row);
        });
    } catch (error) {
        tableBody.innerHTML = '<tr><td colspan="2" class="p-2 text-red-500">Error loading analytics</td></tr>';
    }
    loading.classList.add('hidden');
}

async function fetchFilteredEvents() {
    const loading = document.getElementById('filter-loading');
    const tableBody = document.getElementById('events-body');
    loading.classList.remove('hidden');
    tableBody.innerHTML = '';

    const params = new URLSearchParams();
    const userId = document.getElementById('user_id').value;
    const eventType = document.getElementById('event_type').value;
    const startTime = document.getElementById('start_time').value;
    const endTime = document.getElementById('end_time').value;

    if (userId) params.append('user_id', userId);
    if (eventType) params.append('event_type', eventType);
    if (startTime) params.append('start_time', startTime);
    if (endTime) params.append('end_time', endTime);

    try {
        const response = await fetch(`http://localhost:8080/events/filter?${params}`);
        const data = await response.json();
        data.events.forEach(event => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td class="p-2 border">${event.user_id}</td>
                <td class="p-2 border">${event.event_type}</td>
                <td class="p-2 border">${new Date(event.timestamp).toLocaleString()}</td>
            `;
            tableBody.appendChild(row);
        });
    } catch (error) {
        tableBody.innerHTML = '<tr><td colspan="3" class="p-2 text-red-500">Error loading events</td></tr>';
    }
    loading.classList.add('hidden');
}

// Load analytics on page load
document.addEventListener('DOMContentLoaded', fetchAnalytics);
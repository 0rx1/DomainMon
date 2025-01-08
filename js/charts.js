export function initializeCharts() {
    const tldChart = createChart(document.getElementById('tldChart').getContext('2d'), 'doughnut');
    const similarityChart = createSimilarityChart(document.getElementById('similarityChart').getContext('2d'));

    // Add event listener for similarity threshold changes
    document.getElementById('similarityThreshold').addEventListener('change', (e) => {
        updateSimilarityChart(similarityChart, e.target.value);
    });

    // Initial similarity chart update
    updateSimilarityChart(similarityChart, 70);

    return { tldChart, similarityChart };
}

function createChart(ctx, type) {
    const config = {
        type: type,
        data: {
            labels: [],
            datasets: [{
                data: [],
                backgroundColor: [
                    '#3b82f6', '#10b981', '#f59e0b', '#ef4444',
                    '#8b5cf6', '#ec4899', '#6366f1', '#14b8a6',
                    '#f97316', '#84cc16', '#06b6d4', '#a855f7'
                ],
                borderWidth: 1,
                borderColor: '#fff'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: {
                duration: 2000,
                easing: 'easeOutQuart'
            },
            plugins: {
                legend: {
                    position: type === 'bar' ? 'top' : 'right',
                    labels: {
                        boxWidth: 12,
                        font: {
                            family: "'Inter', sans-serif",
                            weight: 500
                        }
                    }
                }
            },
            layout: {
                padding: {
                    top: 10,
                    right: 10,
                    bottom: 10,
                    left: 10
                }
            }
        }
    };

    // Add specific options for bar chart
    if (type === 'bar') {
        config.options.scales = {
            y: {
                beginAtZero: true,
                title: {
                    display: true,
                    text: 'Number of Domains'
                }
            }
        };
    }

    return new Chart(ctx, config);
}

function createSimilarityChart(ctx) {
    return new Chart(ctx, {
        type: 'radar',
        data: {
            labels: [],
            datasets: [{
                label: 'Similar Domains',
                data: [],
                backgroundColor: 'rgba(59, 130, 246, 0.2)',
                borderColor: '#3b82f6',
                borderWidth: 2,
                pointBackgroundColor: '#3b82f6',
                pointRadius: 4,
                pointHoverRadius: 6,
                similarityData: []
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                r: {
                    beginAtZero: true,
                    ticks: {
                        stepSize: 5
                    },
                    pointLabels: {
                        font: {
                            family: "'Inter', sans-serif",
                            size: 12
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const dataIndex = context.dataIndex;
                            const data = context.dataset.similarityData[dataIndex];
                            if (!data) return 'No data available';
                            
                            return [
                                `Found: ${data.count} similar domains`,
                                `Similarity: ${data.similarity}%`,
                                'Examples:',
                                ...data.examples.map(ex => `• ${ex}`)
                            ];
                        }
                    }
                }
            }
        }
    });
}

async function updateSimilarityChart(chart, threshold) {
    try {
        const response = await fetch(`/api/v1/similarity/${threshold}`, {
            headers: {
                'Accept': 'application/json',
                'Cache-Control': 'no-cache'
            },
            credentials: 'same-origin'
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Failed to fetch similarity data: ${response.status} - ${errorText}`);
        }
        
        const { data } = await response.json();
        if (!data || data.length === 0) {
            throw new Error('No similarity data available');
        }

        // Process data for radar chart
        const processedData = data.map(item => ({
            targetDomain: item.targetDomain.split('.')[0],
            count: item.count,
            similarity: item.similarity,
            examples: item.examples
        }));

        // Update chart data
        chart.data.labels = processedData.map(item => item.targetDomain);
        chart.data.datasets[0].data = processedData.map(item => item.count);
        chart.data.datasets[0].similarityData = processedData;

        // Add details table below chart
        updateSimilarityDetails(processedData);

        chart.update('none');
    } catch (error) {
        console.error('Error updating similarity chart:', error);
        chart.data.labels = ['Error loading data'];
        chart.data.datasets[0].data = [0];
        chart.data.datasets[0].similarityData = [];
        chart.update();
        
        document.getElementById('similarityDetails').innerHTML = `
            <div class="error-message">
                <p><strong>Error:</strong> ${error.message}</p>
                <p>Please try again later or contact support if the issue persists.</p>
            </div>
        `;
    }
}

function updateSimilarityDetails(data) {
    const detailsContainer = document.getElementById('similarityDetails');
    
    const html = `
        <div class="similarity-details">
            <table class="similarity-table">
                <thead>
                    <tr>
                        <th>Target Domain</th>
                        <th>Similar Domains</th>
                        <th>Match %</th>
                        <th>Examples</th>
                    </tr>
                </thead>
                <tbody>
                    ${data.map(item => `
                        <tr>
                            <td>${item.targetDomain}</td>
                            <td>${item.count}</td>
                            <td>${item.similarity}%</td>
                            <td>
                                <div class="examples-list">
                                    ${item.examples.map(ex => `
                                        <span class="example-domain">${ex}</span>
                                    `).join('')}
                                </div>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        </div>
    `;
    
    detailsContainer.innerHTML = html;
}

export function updateCharts(charts, stats) {
    if (!charts || !stats) return;

    try {
        const tldData = Object.entries(stats.domains_per_tld || {});
        const chartData = {
            labels: tldData.map(([tld]) => tld),
            data: tldData.map(([, count]) => count)
        };

        // Store the data for chart type changes
        window.lastChartData = chartData;
        
        updateChart(charts.tldChart, chartData);
    } catch (error) {
        console.error('Error updating chart:', error);
    }
}

function updateChart(chart, { labels, data }) {
    chart.data.labels = labels;
    chart.data.datasets[0].data = data;
    chart.update();
} 
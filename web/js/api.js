const API_URL = "/api";

async function getHello() {
    const response = await fetch(`${API_URL}/hello`);

    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }

    return response.text();
}
async function loadHello() {
    const result = document.getElementById("result");

    try {
        result.textContent = await getHello();
    } catch (error) {
        result.textContent = "Failed to load data";
        console.error(error);
    }
}

loadHello();
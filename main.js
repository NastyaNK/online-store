load = async (url, func) => {
    console.log(url);
    response = await fetch(url);
    let result = await response.text();
    try {
        result = JSON.parse(result);
    } catch {

    }
    console.log(result);
    if (func) func(result);
}

let runners = [];

window.onload = () => {
    runners.forEach((run) => run())
}
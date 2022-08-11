function updateCheck(cb) {
    const xhttp = new XMLHttpRequest();
    xhttp.open(
        'POST',
        '/messen/updateRelevantState?uid=' + cb.id + '&state=' + cb.checked
    );
    cb.checked = !cb.checked;
    xhttp.send();

    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            const children = [...document.getElementsByName(cb.id)];
            children.forEach((child) => {
                child.classList.toggle('text-muted');
                child.classList.toggle('opacity-50');
            });
            cb.checked = !cb.checked;
        }
        if (this.readyState == 4 && this.status == 0) {
            alert(
                'Konnte keine Verbindung zum Server aufbauen!\nStatus bleibt unverändert.'
            );
        }
    };
}

function checkPassword() {
    result = zxcvbn(document.getElementById('pass').value);
    console.log(result.score);
    console.log(document.getElementById('pass').value);
    switch (result.score) {
        case 0:
            document.getElementById('bar').setAttribute('aria-valuenow', '1');
            document.getElementById('bar').className =
                'progress-bar w-1 bg-danger';
            break;
        case 1:
            document.getElementById('bar').setAttribute('aria-valuenow', '25');
            document.getElementById('bar').className =
                'progress-bar w-25 bg-danger';
            break;
        case 2:
            document.getElementById('bar').setAttribute('aria-valuenow', '50');
            document.getElementById('bar').className =
                'progress-bar w-50 bg-warning';
            break;
        case 3:
            document.getElementById('bar').setAttribute('aria-valuenow', '75');
            document.getElementById('bar').className =
                'progress-bar w-75 bg-success';
            break;
        case 4:
            document.getElementById('bar').setAttribute('aria-valuenow', '100');
            document.getElementById('bar').className =
                'progress-bar w-100 bg-success';
            break;

        default:
            break;
    }
}

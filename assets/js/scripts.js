function updateCheck(cb) {   
    const xhttp = new XMLHttpRequest();
    xhttp.open("POST", "/messen/updateRelevantState?uid="+cb.id+"&state="+cb.checked);
    cb.checked = !cb.checked;
    xhttp.send();
   
    xhttp.onreadystatechange = function(){
        if(this.readyState == 4 && this.status==200){
            const children = [...document.getElementsByName(cb.id)];
            children.forEach((child) => { child.classList.toggle("text-muted") });
            cb.checked = !cb.checked;
        }
        if(this.readyState == 4 && this.status==0){
            alert("Konnte keine Verbindung zum Server aufbauen!\nStatus bleibt unverändert.")
        }
    }
}


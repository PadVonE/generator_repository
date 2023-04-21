


function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function updateObjects() {
    i=0;
    while (true) {
        i++

        $(".pers").removeClass("pers")
        $.get("/api/refresh",function (e){
            var html=""
            e.units.forEach(function(unit) {
                console.log(unit.position.Coordinates)
                $(".cord-"+unit.position.Coordinates.X+"-"+unit.position.Coordinates.Y).addClass("pers")
                html +=" <li><div class=\"unit_name\">"+unit.name+"</div> <div class=\"position\"><div class=\"pos_x\">X:"+unit.position.Coordinates.X+"</div><div class=\"pos_y\">Y:"+unit.position.Coordinates.Y+"</div></div></li>"
            })
            $(".ul_units").html

        },"JSON")

        console.log(`Waiting ${i} seconds...`);
        await sleep(i * 50);
    }
}
updateObjects()
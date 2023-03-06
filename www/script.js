const TIMEOUT_LOOP_PICTURES = 15000;
const TIMEOUT_LOOP_EVENTS = 20000;
const TIMEOUT_HIDE_EVENTS = 10000;
const TIMEOUT_LOOP_WEATHER = 30000;
const TIMEOUT_LOOP_UPCOMINGWEATHER = 60000;
const TIMEOUT_HIDE_UPCOMINGWEATHER = 15000;

let viewId = "imageview2"
let eventIndex = 9
let hibernating = false

function httpGetAsync(theUrl, callback) {
    let xmlHttpReq = new XMLHttpRequest();
    xmlHttpReq.onreadystatechange = function () {
        if (xmlHttpReq.readyState === 4 && xmlHttpReq.status === 200)
            callback(xmlHttpReq.responseText);
    }
    xmlHttpReq.open("GET", theUrl, true); // true for asynchronous
    xmlHttpReq.send(null);
}

function httpPostAsync(theUrl, theBody, callback) {
    let xmlHttpReq = new XMLHttpRequest();
    xmlHttpReq.onreadystatechange = function () {
        if (xmlHttpReq.readyState === 4 && xmlHttpReq.status >= 200 && xmlHttpReq.status < 300)
            callback(xmlHttpReq.responseText);
    }
    xmlHttpReq.open("POST", theUrl, true); // true for asynchronous
    xmlHttpReq.send(theBody);
}

function runWithAppStateCheck(appAwareRunnable, hibernateStateCallback) {
    httpGetAsync('/app_state', function (result) {
        const appState = JSON.parse(result)
        const hibernateResult = appState.hibernation.hibernate
        if (hibernateResult !== hibernating) {
            hibernating = hibernateResult
            if (hibernating) {
                document.getElementById("systemlist").innerHTML = '<div><i class="fa-solid fa-xs fa-gears"></i>Hibernating</div>'
            } else {
                document.getElementById("systemlist").innerHTML = ''
            }
        }
        if (!hibernating) {
            appAwareRunnable(appState)
        }
        if (hibernateStateCallback) {
            hibernateStateCallback(hibernating)
        }
    })
}

function padZeroes(input) {
    if (input < 10) {
        return "0" + input
    }
    return input
}

function loadPicture() {
    let newViewId = "imageview"
    if (viewId === "imageview") {
        newViewId = "imageview2"
    }
    const newPictureview = document.getElementById(newViewId)
    newPictureview.style.setProperty("background-image", "url('/picture?t=" + new Date().getTime() + "')")
}

function switchPictures() {
    let newViewId = "imageview"
    if (viewId === "imageview") {
        newViewId = "imageview2"
    }
    let newPictureview = document.getElementById(newViewId)
    let currentPictureview = document.getElementById(viewId)
    currentPictureview.style.removeProperty("background-image")
    currentPictureview.classList.remove("show")
    currentPictureview.classList.add("hide")
    newPictureview.classList.remove("hide")
    newPictureview.classList.add("show")
    viewId = newViewId
}

function loopPictures() {
    window.setTimeout(() => {
        runWithAppStateCheck(() => {
                switchPictures()
                window.setTimeout(loadPicture, 2500)
            }
        )
        loopPictures()
    }, TIMEOUT_LOOP_PICTURES)
}

function loopEvents() {
    window.setTimeout(() => {
        runWithAppStateCheck((appData) => {
            loadEvents(appData)
            window.setTimeout(hideEvents, TIMEOUT_HIDE_EVENTS)
        })
        loopEvents()
    }, TIMEOUT_LOOP_EVENTS)
}

function loadEvents(appData) {
    httpGetAsync('/events', function (result) {
            let eventListContent = ''
            const events = JSON.parse(result);
            eventIndex = (eventIndex + 1) % 10
            const event = events[eventIndex]
            if (event) {
                const date = new Date(event.start);
                let dateString = padZeroes(date.getDate()) + "." + padZeroes(date.getMonth() + 1) + "." + date.getFullYear()
                if (date.getUTCHours() > 0 || date.getUTCMinutes() > 0) {
                    dateString += " " + padZeroes(date.getHours()) + ":" + padZeroes(date.getMinutes())
                }
                let classesSuffix = 'fa-calendar';
                const suffixFromAppState = appData.calendarMappings[event.type];
                if (suffixFromAppState) {
                    classesSuffix = suffixFromAppState;
                }
                const icon = '<i class="fa-solid fa-xs ' + classesSuffix + '"></i>';
                const element = '<div class="eventItem">' + icon + ' <span class="eventDate">' + dateString + '</span> - ' + event.summary + '</div>'
                eventListContent += element
            }
            const eventlist = document.getElementById("eventlist")
            eventlist.innerHTML = eventListContent
            eventlist.classList.remove("hide")
            eventlist.classList.add("show")
        }
    );
}

function hideEvents() {
    const eventlist = document.getElementById("eventlist")
    eventlist.classList.remove("show")
    eventlist.classList.add("hide")
}

function buildWeatherDiv(weather, time, weatherClass) {
    const temp = Math.round(weather.temperature)
    const icon = "https://openweathermap.org/img/wn/" + weather.icon + "@2x.png"
    const precipitation = weather.precipitation
    const wind = Math.round(weather.wind_speed * 100) / 100
    const alert = weather.alert

    const tempDiv = '<div class="temp">' + temp + "°" + '</div>'
    const symbolDiv = '<div class="symbol" style="background-image:url(' + icon + ')"></div>'
    let precipitationDiv = ''
    if (precipitation && precipitation > 0) {
        precipitationDiv = '<div class="precipitation"><i class="fa-solid fa-xs fa-droplet"></i> ' + precipitation +
            '</div>'
    }
    let windDiv = ''
    if (wind && wind > 0) {
        windDiv = '<div class="wind"><i class="fa-solid fa-xs fa-wind"></i> ' + wind + '</div>'
    }
    let alertDiv = ''
    if (alert && alert !== "") {
        alertDiv = '<div class="alert"><i class="fa-solid fa-xs fa-warning"></i> ' + capitalizeFirstLetter(alert) + '</div>'
    }
    let timeDiv = ''
    if (time && time.length > 0) {
        timeDiv = '<div class="time">' + time + '</div>'
    }

    return '<div class="weather ' + weatherClass + '">' + tempDiv + symbolDiv + precipitationDiv + windDiv + alertDiv + timeDiv + '</div>'
}

function loadWeather() {
    httpGetAsync('/weather', function (result) {
        const weather = JSON.parse(result)

        const weatherElementsBefore = document.getElementsByClassName("currentWeather")
        const currentWeatherElement = weatherElementsBefore[0]

        currentWeatherElement.innerHTML = buildWeatherDiv(weather, null, "currentWeather")
    })
}

function loadUpcomingWeather(appState) {
    httpGetAsync('/upcoming_weather', function (result) {
        const weathers = JSON.parse(result)

        let upcomingWeatherElements = ''

        for (let j = 0; j < weathers.length; j++) {
            const weather = weathers[j]

            const time = new Date(weather.time);
            const timeString = padZeroes(time.getHours()) + ":" + padZeroes(time.getMinutes())

            upcomingWeatherElements = upcomingWeatherElements + buildWeatherDiv(weather, timeString, "upcomingWeather", appState)
        }

        const weatherList = document.getElementById('weatherlist')
        weatherList.innerHTML = weatherList.innerHTML + upcomingWeatherElements

        const weatherElements = document.getElementsByClassName("upcomingWeather")
        for (let i = 0; i < weatherElements.length; i++) {
            const weatherElement = weatherElements.item(i)
            weatherElement.classList.remove("hide")
            weatherElement.classList.add("show")
        }
    })
}

function capitalizeFirstLetter(theString) {
    //This takes the first character of the string and capitalizes it
    return theString.substring(0, 1).toUpperCase() + theString.substring(1).toLowerCase();
}

function hideUpcomingWeather() {
    const weatherElements = document.getElementsByClassName("upcomingWeather")
    for (let i = 0; i < weatherElements.length; i++) {
        const weatherElement = weatherElements.item(i)
        weatherElement.classList.remove("show")
        weatherElement.classList.add("hide")
    }
    window.setTimeout(removeUpcomingWeather, 2500)
}

function removeUpcomingWeather() {
    const upcomingWeatherElements = document.getElementsByClassName("upcomingWeather")
    for (let i = upcomingWeatherElements.length - 1; i >= 0; i--) {
        const weatherElement = upcomingWeatherElements.item(i)
        weatherElement.remove()
    }
}

function loopWeather() {
    window.setTimeout(() => {
        runWithAppStateCheck(() => {
            loadWeather()
        })
        loopWeather()
    }, TIMEOUT_LOOP_WEATHER)
}


function loopUpcomingWeather() {
    window.setTimeout(() => {
        runWithAppStateCheck(() => {
            loadUpcomingWeather()
            window.setTimeout(hideUpcomingWeather, TIMEOUT_HIDE_UPCOMINGWEATHER)
        })
        loopUpcomingWeather()
    }, TIMEOUT_LOOP_UPCOMINGWEATHER)
}

function init_pictureframe() {
    loadPicture()
    loadPicture()
    runWithAppStateCheck((appData) => {
        loadEvents(appData)
    })
    loadWeather()
    loadUpcomingWeather()
    window.setTimeout(hideUpcomingWeather, TIMEOUT_HIDE_UPCOMINGWEATHER)
    loopPictures()
    loopEvents()
    loopWeather()
    loopUpcomingWeather()
}

function admin_loadHibernateState() {
    runWithAppStateCheck(() => {
    }, (hibernateState) => {
        const hibernateStateSpan = document.getElementById("hibernateState")
        hibernateStateSpan.innerText = hibernateState
    })
}

function admin_toggleHibernateState() {
    runWithAppStateCheck(() => {
        }, (hibernateState) => {
            const newState = !hibernateState
            const hibernateBody = {hibernate: newState}
            httpPostAsync('/hibernate', JSON.stringify(hibernateBody), admin_loadHibernateState)
        }
    )
}

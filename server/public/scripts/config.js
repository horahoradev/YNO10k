
let config = {
	chat: {
    	name: '',
		trip: ''
	}
  };

  
  function saveChatConfig () {
    config.chat.name = nameInput.value;
	config.chat.trip = tripInput.value;
    updateConfig(config);
  };
  
  function loadOrInitConfig() {
	let savedConfig = config;
	let configjson;
	try {
		configjson = JSON.stringify(config);
		if (!window.localStorage.hasOwnProperty('config')) {
    		window.localStorage.setItem('config', configjson);
		}
    	else {
    		savedConfig = JSON.parse(window.localStorage.getItem('config'));
				nameInput.value = savedConfig.chat.name;
				tripInput.value = savedConfig.chat.trip;
    	}
	} catch(e) {
		console.error(e);
		console.log("Something went wrong when loading your saved configurations. Your configs will be overwritten.");
		console.log("Your old configs: " + window.localStorage.getItem('config'));
    	window.localStorage.setItem('config', configjson);
	}
	
	config = savedConfig;
  }
  
  function updateConfig(config) {
    try {
      window.localStorage.config = JSON.stringify(config);
    } catch (error) {
		PrintChatInfo("Something went wrong when saving your configurations.")
    }
  }


function getProfileConfigName() {
	if(window.localStorage.hasOwnProperty('config')) {
		var savedConfig = JSON.parse(window.localStorage.getItem('config'));
		return savedConfig.chat.name;
	} else {
		return "";
	}
}

function getProfileConfigTrip() {
	if(window.localStorage.hasOwnProperty('config')) {
		var savedConfig = JSON.parse(window.localStorage.getItem('config'));
		return savedConfig.chat.trip;
	} else {
		return "";
	}
}
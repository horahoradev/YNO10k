
let switchSyncWhiteLists = {
	nikki: [
		142, //rain
		143, //snow
		250, //masada's ship is crashing
		251	//masada's ship is crashed
		]
}

function InitSwitchSyncWhiteLists() {
	if(switchSyncWhiteLists[gameName]) {
		if(switchSyncWhiteLists[gameName].length) {
			Module._SetSwitchSync(1);
			for(let id of switchSyncWhiteLists[gameName]) {
				Module._SetSwitchSyncWhiteList(id, 1);
			}
		}
	}
}
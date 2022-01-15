
function SetPlayersVolumeCommand(args) {
	if(args.length == 2) {
		let vol = parseInt(args[1]);
		if(vol >= 0 && vol <= 100) {
			Module._SetPlayersVolume(vol);
			return true;
		}
	}
	return false;
}
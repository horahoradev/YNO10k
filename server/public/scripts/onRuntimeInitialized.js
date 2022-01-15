function onRuntimeInitialized() {
	let host = Module.allocate(Module.intArrayFromString(WSAddress), Module.ALLOC_NORMAL);
	Module._SetWSHost(host);
	Module._free(host);

	InitSwitchSyncWhiteLists();
	initChat();
}

Module['onRuntimeInitialized'] = onRuntimeInitialized;
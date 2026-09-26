.PHONY: smoke load stress soak

smoke:
	k6 run tools/k6/scenarios/smoke.js

load:
	k6 run tools/k6/scenarios/load.js

stress:
	k6 run tools/k6/scenarios/stress.js

soak:
	k6 run tools/k6/scenarios/soak.js
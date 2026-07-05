export namespace main {
	
	export class TestSessionParams {
	    repeatCount: number;
	    stepIntervalSeconds: number;
	    outputDir: string;
	    outputFileName: string;
	
	    static createFrom(source: any = {}) {
	        return new TestSessionParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repeatCount = source["repeatCount"];
	        this.stepIntervalSeconds = source["stepIntervalSeconds"];
	        this.outputDir = source["outputDir"];
	        this.outputFileName = source["outputFileName"];
	    }
	}

}


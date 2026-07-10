export namespace main {
	
	export class CaptureRegionInput {
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new CaptureRegionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}
	export class ClickPointInput {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new ClickPointInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class CaptureSessionParams {
	    repeatCount: number;
	    stepIntervalSeconds: number;
	    outputDir: string;
	    outputFileName: string;
	    captureRegion: CaptureRegionInput;
	    advanceClickPoint: ClickPointInput;
	
	    static createFrom(source: any = {}) {
	        return new CaptureSessionParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repeatCount = source["repeatCount"];
	        this.stepIntervalSeconds = source["stepIntervalSeconds"];
	        this.outputDir = source["outputDir"];
	        this.outputFileName = source["outputFileName"];
	        this.captureRegion = this.convertValues(source["captureRegion"], CaptureRegionInput);
	        this.advanceClickPoint = this.convertValues(source["advanceClickPoint"], ClickPointInput);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}


export namespace main {
	
	export class HideoutLevelDTO {
	    id: string;
	    level: number;
	    completed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HideoutLevelDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.level = source["level"];
	        this.completed = source["completed"];
	    }
	}
	export class HideoutStationDTO {
	    id: string;
	    name: string;
	    levels: HideoutLevelDTO[];
	
	    static createFrom(source: any = {}) {
	        return new HideoutStationDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.levels = this.convertValues(source["levels"], HideoutLevelDTO);
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
	export class ItemRequirementDTO {
	    id: string;
	    name: string;
	    iconPath: string;
	    hideoutTotalFIR: number;
	    hideoutUsedFIR: number;
	    hideoutTotalNorm: number;
	    hideoutUsedNorm: number;
	    questTotalFIR: number;
	    questTotalNorm: number;
	    stashFIR: number;
	    stashNorm: number;
	
	    static createFrom(source: any = {}) {
	        return new ItemRequirementDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.iconPath = source["iconPath"];
	        this.hideoutTotalFIR = source["hideoutTotalFIR"];
	        this.hideoutUsedFIR = source["hideoutUsedFIR"];
	        this.hideoutTotalNorm = source["hideoutTotalNorm"];
	        this.hideoutUsedNorm = source["hideoutUsedNorm"];
	        this.questTotalFIR = source["questTotalFIR"];
	        this.questTotalNorm = source["questTotalNorm"];
	        this.stashFIR = source["stashFIR"];
	        this.stashNorm = source["stashNorm"];
	    }
	}
	export class QuestItemDTO {
	    name: string;
	    quantity: number;
	    foundInRaid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QuestItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.quantity = source["quantity"];
	        this.foundInRaid = source["foundInRaid"];
	    }
	}
	export class QuestDTO {
	    id: string;
	    name: string;
	    trader: string;
	    items: QuestItemDTO[];
	
	    static createFrom(source: any = {}) {
	        return new QuestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.trader = source["trader"];
	        this.items = this.convertValues(source["items"], QuestItemDTO);
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
	
	export class StashItemDTO {
	    name: string;
	    id: string;
	    quantity: number;
	    fir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new StashItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.id = source["id"];
	        this.quantity = source["quantity"];
	        this.fir = source["fir"];
	    }
	}
	export class StashResultDTO {
	    items: StashItemDTO[];
	    unmatched: StashItemDTO[];
	    imageCount: number;
	
	    static createFrom(source: any = {}) {
	        return new StashResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], StashItemDTO);
	        this.unmatched = this.convertValues(source["unmatched"], StashItemDTO);
	        this.imageCount = source["imageCount"];
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
	export class SyncInfoDTO {
	    itemCount: number;
	    newItems: number;
	    needsIcons: boolean;
	    missingIcons: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncInfoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemCount = source["itemCount"];
	        this.newItems = source["newItems"];
	        this.needsIcons = source["needsIcons"];
	        this.missingIcons = source["missingIcons"];
	    }
	}

}


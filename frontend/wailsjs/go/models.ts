export namespace main {
	
	export class WalletStatus {
	    exists: boolean;
	    locked: boolean;
	    hasPassword: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WalletStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exists = source["exists"];
	        this.locked = source["locked"];
	        this.hasPassword = source["hasPassword"];
	    }
	}

}


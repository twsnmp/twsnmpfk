export namespace datastore {
	
	export class MapConfEnt {
	    MapName: string;
	    PollInt: number;
	    Timeout: number;
	    Retry: number;
	    LogDays: number;
	    SnmpMode: string;
	    Community: string;
	    SnmpUser: string;
	    SnmpPassword: string;
	    EnableSyslogd: boolean;
	    EnableTrapd: boolean;
	    EnableArpWatch: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MapConfEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.MapName = source["MapName"];
	        this.PollInt = source["PollInt"];
	        this.Timeout = source["Timeout"];
	        this.Retry = source["Retry"];
	        this.LogDays = source["LogDays"];
	        this.SnmpMode = source["SnmpMode"];
	        this.Community = source["Community"];
	        this.SnmpUser = source["SnmpUser"];
	        this.SnmpPassword = source["SnmpPassword"];
	        this.EnableSyslogd = source["EnableSyslogd"];
	        this.EnableTrapd = source["EnableTrapd"];
	        this.EnableArpWatch = source["EnableArpWatch"];
	    }
	}

}

export namespace main {
	
	export class Settings {
	    Kiosk: boolean;
	    Lock: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Kiosk = source["Kiosk"];
	        this.Lock = source["Lock"];
	    }
	}

}


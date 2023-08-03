export namespace datastore {
	
	export class AIConfEnt {
	    HighThreshold: number;
	    LowThreshold: number;
	    WarnThreshold: number;
	
	    static createFrom(source: any = {}) {
	        return new AIConfEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.HighThreshold = source["HighThreshold"];
	        this.LowThreshold = source["LowThreshold"];
	        this.WarnThreshold = source["WarnThreshold"];
	    }
	}
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
	export class NotifyConfEnt {
	    MailServer: string;
	    InsecureSkipVerify: boolean;
	    User: string;
	    Password: string;
	    MailTo: string;
	    MailFrom: string;
	    Subject: string;
	    Interval: number;
	    Level: string;
	    Report: boolean;
	    NotifyRepair: boolean;
	    ExecCmd: string;
	
	    static createFrom(source: any = {}) {
	        return new NotifyConfEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.MailServer = source["MailServer"];
	        this.InsecureSkipVerify = source["InsecureSkipVerify"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.MailTo = source["MailTo"];
	        this.MailFrom = source["MailFrom"];
	        this.Subject = source["Subject"];
	        this.Interval = source["Interval"];
	        this.Level = source["Level"];
	        this.Report = source["Report"];
	        this.NotifyRepair = source["NotifyRepair"];
	        this.ExecCmd = source["ExecCmd"];
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


export namespace bridge {
	
	export class PythonEnvReport {
	    executable: string;
	    version: string;
	    mode: string;
	    missing_deps: string[];
	    app_load_error: string;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PythonEnvReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.executable = source["executable"];
	        this.version = source["version"];
	        this.mode = source["mode"];
	        this.missing_deps = source["missing_deps"];
	        this.app_load_error = source["app_load_error"];
	        this.success = source["success"];
	    }
	}

}

export namespace models {
	
	export class ActionResult {
	    success: boolean;
	    message?: string;
	    sha?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.sha = source["sha"];
	        this.error = source["error"];
	    }
	}
	export class CommitItem {
	    sha: string;
	    message: string;
	    author: string;
	    date: string;
	    tag?: string;
	
	    static createFrom(source: any = {}) {
	        return new CommitItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sha = source["sha"];
	        this.message = source["message"];
	        this.author = source["author"];
	        this.date = source["date"];
	        this.tag = source["tag"];
	    }
	}
	export class CommitProposal {
	    success: boolean;
	    error?: string;
	    type: string;
	    scope?: string;
	    subject: string;
	    body?: string;
	    raw_formatted: string;
	
	    static createFrom(source: any = {}) {
	        return new CommitProposal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.type = source["type"];
	        this.scope = source["scope"];
	        this.subject = source["subject"];
	        this.body = source["body"];
	        this.raw_formatted = source["raw_formatted"];
	    }
	}
	export class CommitsResponse {
	    success: boolean;
	    error?: string;
	    commits: CommitItem[];
	
	    static createFrom(source: any = {}) {
	        return new CommitsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.commits = this.convertValues(source["commits"], CommitItem);
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
	export class ConfigResponse {
	    success: boolean;
	    error?: string;
	    ollama_base_url: string;
	    ollama_model: string;
	    timeout_s: number;
	    mock_interface: boolean;
	    mock_ai: boolean;
	    docker_llm: boolean;
	    enable_notifications: boolean;
	    python_bin?: string;
	    resolved_python_bin?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.ollama_base_url = source["ollama_base_url"];
	        this.ollama_model = source["ollama_model"];
	        this.timeout_s = source["timeout_s"];
	        this.mock_interface = source["mock_interface"];
	        this.mock_ai = source["mock_ai"];
	        this.docker_llm = source["docker_llm"];
	        this.enable_notifications = source["enable_notifications"];
	        this.python_bin = source["python_bin"];
	        this.resolved_python_bin = source["resolved_python_bin"];
	    }
	}
	export class DiffFile {
	    path: string;
	    status: string;
	    added: number;
	    removed: number;
	    binary: boolean;
	    patch: string;
	    has_secrets_masked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiffFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.status = source["status"];
	        this.added = source["added"];
	        this.removed = source["removed"];
	        this.binary = source["binary"];
	        this.patch = source["patch"];
	        this.has_secrets_masked = source["has_secrets_masked"];
	    }
	}
	export class DiffResponse {
	    success: boolean;
	    error?: string;
	    files: DiffFile[];
	    staged_count: number;
	
	    static createFrom(source: any = {}) {
	        return new DiffResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.files = this.convertValues(source["files"], DiffFile);
	        this.staged_count = source["staged_count"];
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
	export class DockerLogsResponse {
	    success: boolean;
	    logs: string;
	    tail_lines: number;
	    running: boolean;
	    container_id?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DockerLogsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.logs = source["logs"];
	        this.tail_lines = source["tail_lines"];
	        this.running = source["running"];
	        this.container_id = source["container_id"];
	        this.error = source["error"];
	    }
	}
	export class DockerStatusResponse {
	    success: boolean;
	    running: boolean;
	    container_id?: string;
	    state?: string;
	    health?: string;
	    status_text?: string;
	    is_ready: boolean;
	    is_pulling: boolean;
	    pull_info?: string;
	    target_model?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DockerStatusResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.running = source["running"];
	        this.container_id = source["container_id"];
	        this.state = source["state"];
	        this.health = source["health"];
	        this.status_text = source["status_text"];
	        this.is_ready = source["is_ready"];
	        this.is_pulling = source["is_pulling"];
	        this.pull_info = source["pull_info"];
	        this.target_model = source["target_model"];
	        this.error = source["error"];
	    }
	}
	export class PingResponse {
	    success: boolean;
	    reachable: boolean;
	    latency_ms: number;
	    installed_models: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new PingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.reachable = source["reachable"];
	        this.latency_ms = source["latency_ms"];
	        this.installed_models = source["installed_models"];
	        this.error = source["error"];
	    }
	}
	export class ReleaseNotesResponse {
	    success: boolean;
	    error?: string;
	    from_tag: string;
	    to_tag: string;
	    markdown: string;
	    commit_count: number;
	
	    static createFrom(source: any = {}) {
	        return new ReleaseNotesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.from_tag = source["from_tag"];
	        this.to_tag = source["to_tag"];
	        this.markdown = source["markdown"];
	        this.commit_count = source["commit_count"];
	    }
	}
	export class RepoStatus {
	    is_git_repo: boolean;
	    repo_path: string;
	    branch: string;
	    staged_count: number;
	    unstaged_count: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_git_repo = source["is_git_repo"];
	        this.repo_path = source["repo_path"];
	        this.branch = source["branch"];
	        this.staged_count = source["staged_count"];
	        this.unstaged_count = source["unstaged_count"];
	        this.error = source["error"];
	    }
	}

}


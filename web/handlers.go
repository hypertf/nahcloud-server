package web

import (
	"encoding/base64"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hypertf/nahcloud/domain"
	"github.com/hypertf/nahcloud/service"
	"github.com/hypertf/nahcloud/web/static"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// ServeLogo serves the logo image
func (h *Handler) ServeLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Write(static.Logo)
}

// Dashboard shows the main dashboard
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>NahCloud Console</title>
    <script src="https://unpkg.com/htmx.org@1.9.6"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>{{.CSS}}</style>
</head>
<body class="bg-slate-50 text-slate-800 font-sans min-h-screen">
    <div class="flex min-h-screen">
        <aside class="w-60 bg-white border-r border-slate-200 py-6 fixed h-screen overflow-y-auto">
            <div class="px-6 pb-6 border-b border-slate-200 mb-4">
                <div class="flex items-center gap-3">
                    <img src="/web/static/logo.png" alt="NahCloud" class="w-10 h-10 rounded-lg">
                    <div>
                        <h1 class="text-xl font-bold text-[#2878B5]">NahCloud</h1>
                        <span class="text-xs text-slate-500 font-medium">Console</span>
                    </div>
                </div>
            </div>
            <nav class="px-3">
                <a href="#" hx-get="/web/projects" hx-target="#content" class="flex items-center gap-3 px-4 py-3 text-slate-500 rounded-lg font-medium text-sm hover:bg-slate-50 hover:text-slate-800 transition-all mb-1">
                    <svg class="w-5 h-5 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"></path>
                    </svg>
                    Projects
                </a>
                <a href="#" hx-get="/web/instances" hx-target="#content" class="flex items-center gap-3 px-4 py-3 text-slate-500 rounded-lg font-medium text-sm hover:bg-slate-50 hover:text-slate-800 transition-all mb-1">
                    <svg class="w-5 h-5 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"></path>
                    </svg>
                    Instances
                </a>
                <a href="#" hx-get="/web/metadata" hx-target="#content" class="flex items-center gap-3 px-4 py-3 text-slate-500 rounded-lg font-medium text-sm hover:bg-slate-50 hover:text-slate-800 transition-all mb-1">
                    <svg class="w-5 h-5 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"></path>
                    </svg>
                    Metadata
                </a>
                <a href="#" hx-get="/web/storage" hx-target="#content" class="flex items-center gap-3 px-4 py-3 text-slate-500 rounded-lg font-medium text-sm hover:bg-slate-50 hover:text-slate-800 transition-all mb-1">
                    <svg class="w-5 h-5 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"></path>
                    </svg>
                    Storage
                </a>
            </nav>
        </aside>
        <main class="flex-1 ml-60 p-8">
            <div id="content" class="max-w-6xl">
                <div class="bg-[#2878B5] text-white p-10 rounded-xl text-center">
                    <h2 class="text-2xl font-semibold mb-2">Welcome to NahCloud</h2>
                    <p class="opacity-90">Select a resource type from the sidebar to get started.</p>
                </div>
            </div>
        </main>
    </div>
</body>
</html>
`
	data := struct {
		CSS template.CSS
	}{
		CSS: template.CSS(static.CSS),
	}
	t := template.Must(template.New("dashboard").Parse(tmpl))
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

// Projects handlers
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(domain.ProjectListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
        <h2 class="text-lg font-semibold">Projects</h2>
        <button class="btn btn-primary" hx-get="/web/projects/new" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
            </svg>
            New Project
        </button>
    </div>
    <table class="w-full">
        <thead>
            <tr>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">ID</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Name</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Created At</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr class="hover:bg-slate-50">
                <td class="px-6 py-4 border-b border-slate-100 font-mono text-sm text-slate-500">{{.ID}}</td>
                <td class="px-6 py-4 border-b border-slate-100 font-medium">{{.Name}}</td>
                <td class="px-6 py-4 border-b border-slate-100 text-slate-500">{{.CreatedAt.Format "2006-01-02 15:04:05"}}</td>
                <td class="px-6 py-4 border-b border-slate-100">
                    <div class="flex gap-2">
                        <button class="btn btn-secondary btn-sm" hx-get="/web/projects/{{.ID}}/edit" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">Edit</button>
                        <button class="btn btn-danger btn-sm" hx-delete="/web/projects/{{.ID}}" hx-target="closest tr" hx-confirm="Are you sure you want to delete this project?">Delete</button>
                    </div>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<div id="modal" class="hidden fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm items-start justify-center" onclick="if(event.target === this) this.style.display='none'">
    <div class="bg-white rounded-xl shadow-xl w-full max-w-lg mt-[10vh] h-fit overflow-hidden" onclick="event.stopPropagation()">
        <div id="modal-content"></div>
    </div>
</div>

<style>
#modal[style*="block"] { display: flex !important; }
</style>
`

	t := template.Must(template.New("projects").Parse(tmpl))
	if err := t.Execute(w, projects); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewProjectForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">New Project</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<form hx-post="/web/projects" hx-target="#content" hx-on::after-request="if(event.detail.successful) document.getElementById('modal').style.display='none'">
    <div class="p-6">
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="name">Project Name</label>
            <input type="text" id="name" name="name" placeholder="Enter project name" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
    </div>
    <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3 bg-slate-50">
        <button type="button" class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
        <button type="submit" class="btn btn-primary">Create Project</button>
    </div>
</form>
`
	w.Write([]byte(tmpl))
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := domain.CreateProjectRequest{
		Name: r.FormValue("name"),
	}

	_, err := h.service.CreateProject(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.ListProjects(w, r)
}

func (h *Handler) EditProjectForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	project, err := h.service.GetProject(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">Edit Project</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<form hx-put="/web/projects/{{.ID}}" hx-target="#content" hx-on::after-request="if(event.detail.successful) document.getElementById('modal').style.display='none'">
    <div class="p-6">
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="name">Project Name</label>
            <input type="text" id="name" name="name" value="{{.Name}}" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
    </div>
    <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3 bg-slate-50">
        <button type="button" class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
        <button type="submit" class="btn btn-primary">Update Project</button>
    </div>
</form>
`
	t := template.Must(template.New("edit-project").Parse(tmpl))
	if err := t.Execute(w, project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := domain.UpdateProjectRequest{
		Name: r.FormValue("name"),
	}

	_, err := h.service.UpdateProject(id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.ListProjects(w, r)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteProject(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Instances handlers
func (h *Handler) ListInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := h.service.ListInstances(domain.InstanceListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects, err := h.service.ListProjects(domain.ProjectListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
        <h2 class="text-lg font-semibold">Instances</h2>
        <button class="btn btn-primary" hx-get="/web/instances/new" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
            </svg>
            New Instance
        </button>
    </div>
    <table class="w-full">
        <thead>
            <tr>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">ID</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Project</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Name</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">CPU</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Memory</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Image</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Status</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .Instances}}
            <tr class="hover:bg-slate-50">
                <td class="px-6 py-4 border-b border-slate-100 font-mono text-sm text-slate-500">{{.ID}}</td>
                <td class="px-6 py-4 border-b border-slate-100 font-mono text-sm text-slate-500">{{.ProjectID}}</td>
                <td class="px-6 py-4 border-b border-slate-100 font-medium">{{.Name}}</td>
                <td class="px-6 py-4 border-b border-slate-100">{{.CPU}} vCPU</td>
                <td class="px-6 py-4 border-b border-slate-100">{{.MemoryMB}} MB</td>
                <td class="px-6 py-4 border-b border-slate-100"><code class="bg-slate-100 px-2 py-0.5 rounded text-sm">{{.Image}}</code></td>
                <td class="px-6 py-4 border-b border-slate-100">
                    {{if eq .Status "running"}}
                    <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-50 text-emerald-600">
                        <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                        Running
                    </span>
                    {{else}}
                    <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-red-50 text-red-600">
                        <span class="w-1.5 h-1.5 rounded-full bg-red-500"></span>
                        Stopped
                    </span>
                    {{end}}
                </td>
                <td class="px-6 py-4 border-b border-slate-100">
                    <div class="flex gap-2">
                        <button class="btn btn-secondary btn-sm" hx-get="/web/instances/{{.ID}}/edit" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">Edit</button>
                        <button class="btn btn-danger btn-sm" hx-delete="/web/instances/{{.ID}}" hx-target="closest tr" hx-confirm="Are you sure you want to delete this instance?">Delete</button>
                    </div>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<div id="modal" class="hidden fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm items-start justify-center" onclick="if(event.target === this) this.style.display='none'">
    <div class="bg-white rounded-xl shadow-xl w-full max-w-lg mt-[10vh] h-fit overflow-hidden" onclick="event.stopPropagation()">
        <div id="modal-content"></div>
    </div>
</div>

<style>
#modal[style*="block"] { display: flex !important; }
</style>
`

	data := struct {
		Instances []*domain.Instance
		Projects  []*domain.Project
	}{
		Instances: instances,
		Projects:  projects,
	}

	t := template.Must(template.New("instances").Parse(tmpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewInstanceForm(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(domain.ProjectListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">New Instance</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<form hx-post="/web/instances" hx-target="#content" hx-on::after-request="if(event.detail.successful) document.getElementById('modal').style.display='none'">
    <div class="p-6">
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="project_id">Project</label>
            <select id="project_id" name="project_id" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all bg-white">
                <option value="">Select a project</option>
                {{range .}}
                <option value="{{.ID}}">{{.Name}}</option>
                {{end}}
            </select>
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="name">Instance Name</label>
            <input type="text" id="name" name="name" placeholder="my-instance" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
        <div class="grid grid-cols-2 gap-4 mb-5">
            <div>
                <label class="block text-sm font-medium mb-1.5" for="cpu">CPU (vCPU)</label>
                <input type="number" id="cpu" name="cpu" value="1" min="1" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
            </div>
            <div>
                <label class="block text-sm font-medium mb-1.5" for="memory_mb">Memory (MB)</label>
                <input type="number" id="memory_mb" name="memory_mb" value="512" min="128" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
            </div>
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="image">Image</label>
            <input type="text" id="image" name="image" value="ubuntu:20.04" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="status">Initial Status</label>
            <select id="status" name="status" class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all bg-white">
                <option value="running">Running</option>
                <option value="stopped">Stopped</option>
            </select>
        </div>
    </div>
    <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3 bg-slate-50">
        <button type="button" class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
        <button type="submit" class="btn btn-primary">Create Instance</button>
    </div>
</form>
`
	t := template.Must(template.New("new-instance").Parse(tmpl))
	if err := t.Execute(w, projects); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cpu, _ := strconv.Atoi(r.FormValue("cpu"))
	memoryMB, _ := strconv.Atoi(r.FormValue("memory_mb"))

	req := domain.CreateInstanceRequest{
		ProjectID: r.FormValue("project_id"),
		Name:      r.FormValue("name"),
		CPU:       cpu,
		MemoryMB:  memoryMB,
		Image:     r.FormValue("image"),
		Status:    r.FormValue("status"),
	}

	_, err := h.service.CreateInstance(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.ListInstances(w, r)
}

func (h *Handler) EditInstanceForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	instance, err := h.service.GetInstance(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	projects, err := h.service.ListProjects(domain.ProjectListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">Edit Instance</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<form hx-put="/web/instances/{{.Instance.ID}}" hx-target="#content" hx-on::after-request="if(event.detail.successful) document.getElementById('modal').style.display='none'">
    <div class="p-6">
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="project_id">Project</label>
            <select id="project_id" name="project_id" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all bg-white">
                {{range .Projects}}
                <option value="{{.ID}}" {{if eq .ID $.Instance.ProjectID}}selected{{end}}>{{.Name}}</option>
                {{end}}
            </select>
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="name">Instance Name</label>
            <input type="text" id="name" name="name" value="{{.Instance.Name}}" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
        <div class="grid grid-cols-2 gap-4 mb-5">
            <div>
                <label class="block text-sm font-medium mb-1.5" for="cpu">CPU (vCPU)</label>
                <input type="number" id="cpu" name="cpu" value="{{.Instance.CPU}}" min="1" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
            </div>
            <div>
                <label class="block text-sm font-medium mb-1.5" for="memory_mb">Memory (MB)</label>
                <input type="number" id="memory_mb" name="memory_mb" value="{{.Instance.MemoryMB}}" min="128" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
            </div>
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="image">Image</label>
            <input type="text" id="image" name="image" value="{{.Instance.Image}}" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="status">Status</label>
            <select id="status" name="status" class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all bg-white">
                <option value="running" {{if eq .Instance.Status "running"}}selected{{end}}>Running</option>
                <option value="stopped" {{if eq .Instance.Status "stopped"}}selected{{end}}>Stopped</option>
            </select>
        </div>
    </div>
    <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3 bg-slate-50">
        <button type="button" class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
        <button type="submit" class="btn btn-primary">Update Instance</button>
    </div>
</form>
`

	data := struct {
		Instance domain.Instance
		Projects []*domain.Project
	}{
		Instance: *instance,
		Projects: projects,
	}

	t := template.Must(template.New("edit-instance").Parse(tmpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	cpuStr := r.FormValue("cpu")
	cpu, _ := strconv.Atoi(cpuStr)
	memoryMBStr := r.FormValue("memory_mb")
	memoryMB, _ := strconv.Atoi(memoryMBStr)
	image := r.FormValue("image")
	status := r.FormValue("status")

	req := domain.UpdateInstanceRequest{
		Name:     &name,
		CPU:      &cpu,
		MemoryMB: &memoryMB,
		Image:    &image,
		Status:   &status,
	}

	_, err := h.service.UpdateInstance(id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.ListInstances(w, r)
}

func (h *Handler) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteInstance(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Metadata handlers
func (h *Handler) ListMetadata(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	paths, err := h.service.ListMetadata(domain.MetadataListOptions{Prefix: prefix})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var metadata []domain.Metadata
	for _, meta := range paths {
		metadata = append(metadata, *meta)
	}

	tmpl := `
<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
        <h2 class="text-lg font-semibold">Metadata</h2>
        <button class="btn btn-primary" hx-get="/web/metadata/new" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
            </svg>
            New Metadata
        </button>
    </div>
    <div class="px-6 py-4 border-b border-slate-200">
        <div class="max-w-sm">
            <label class="block text-sm font-medium mb-1.5" for="prefix-filter">Filter by prefix</label>
            <input type="text" id="prefix-filter" name="prefix" hx-get="/web/metadata" hx-target="#content" hx-trigger="input changed delay:500ms" value="{{.Prefix}}" placeholder="Enter prefix to filter..." class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
    </div>
    <table class="w-full">
        <thead>
            <tr>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Path</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Value</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Updated At</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .Metadata}}
            <tr class="hover:bg-slate-50" id="row-{{.ID}}">
                <td class="px-6 py-4 border-b border-slate-100"><code class="bg-slate-100 px-2 py-0.5 rounded text-sm">{{.Path}}</code></td>
                <td class="px-6 py-4 border-b border-slate-100 max-w-xs truncate">{{.Value}}</td>
                <td class="px-6 py-4 border-b border-slate-100 text-slate-500">{{.UpdatedAt.Format "2006-01-02 15:04:05"}}</td>
                <td class="px-6 py-4 border-b border-slate-100">
                    <div class="flex gap-2">
                        <button class="btn btn-secondary btn-sm" hx-get="/web/metadata/edit?path={{.Path}}" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">Edit</button>
                        <button class="btn btn-danger btn-sm" onclick="showDeleteConfirm('{{.Path}}', 'row-{{.ID}}')">Delete</button>
                    </div>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<div id="modal" class="hidden fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm items-start justify-center" onclick="if(event.target === this) this.style.display='none'">
    <div class="bg-white rounded-xl shadow-xl w-full max-w-lg mt-[10vh] h-fit overflow-hidden" onclick="event.stopPropagation()">
        <div id="modal-content"></div>
    </div>
</div>

<div id="confirm-modal" class="hidden fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm items-start justify-center" onclick="if(event.target === this) hideDeleteConfirm()">
    <div class="bg-white rounded-xl shadow-xl w-full max-w-sm mt-[20vh] h-fit overflow-hidden" onclick="event.stopPropagation()">
        <div class="p-5">
            <h3 class="text-base font-semibold text-slate-800 mb-2">Delete Metadata</h3>
            <p class="text-sm text-slate-600">Delete <code id="confirm-path" class="bg-slate-100 px-1.5 py-0.5 rounded text-xs"></code>?</p>
        </div>
        <div class="px-5 py-3 border-t border-slate-200 flex justify-end gap-2 bg-slate-50">
            <button class="btn btn-secondary btn-sm" onclick="hideDeleteConfirm()">Cancel</button>
            <button id="confirm-delete-btn" class="btn btn-danger btn-sm">Delete</button>
        </div>
    </div>
</div>

<style>
#modal[style*="block"] { display: flex !important; }
#confirm-modal[style*="block"] { display: flex !important; }
</style>

<script>
function showDeleteConfirm(path, rowId) {
    document.getElementById('confirm-path').textContent = path;
    document.getElementById('confirm-modal').style.display = 'block';
    document.getElementById('confirm-delete-btn').onclick = function() {
        htmx.ajax('DELETE', '/web/metadata/delete?path=' + encodeURIComponent(path), {target: '#' + rowId, swap: 'outerHTML'});
        hideDeleteConfirm();
    };
}
function hideDeleteConfirm() {
    document.getElementById('confirm-modal').style.display = 'none';
}
</script>
`

	data := struct {
		Metadata []domain.Metadata
		Prefix   string
	}{
		Metadata: metadata,
		Prefix:   prefix,
	}

	t := template.Must(template.New("metadata").Parse(tmpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewMetadataForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">New Metadata</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<form hx-post="/web/metadata" hx-target="#content" hx-on::after-request="if(event.detail.successful) document.getElementById('modal').style.display='none'">
    <div class="p-6">
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="path">Path</label>
            <input type="text" id="path" name="path" placeholder="config/settings/key" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="value">Value</label>
            <textarea id="value" name="value" rows="4" placeholder="Enter value..." required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all resize-none"></textarea>
        </div>
    </div>
    <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3 bg-slate-50">
        <button type="button" class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
        <button type="submit" class="btn btn-primary">Create Metadata</button>
    </div>
</form>
`
	w.Write([]byte(tmpl))
}

func (h *Handler) CreateMetadata(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	path := r.FormValue("path")
	value := r.FormValue("value")

	if _, err := h.service.CreateMetadata(domain.CreateMetadataRequest{Path: path, Value: value}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.ListMetadata(w, r)
}

func (h *Handler) EditMetadataForm(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	metadata, err := h.service.GetMetadataByPath(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">Edit Metadata</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<form hx-put="/web/metadata/update" hx-target="#content" hx-on::after-request="if(event.detail.successful) document.getElementById('modal').style.display='none'">
    <input type="hidden" name="id" value="{{.ID}}">
    <div class="p-6">
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="path">Path</label>
            <input type="text" id="path" name="path" value="{{.Path}}" readonly class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg bg-slate-50 text-slate-500">
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="value">Value</label>
            <textarea id="value" name="value" rows="4" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all resize-none">{{.Value}}</textarea>
        </div>
    </div>
    <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3 bg-slate-50">
        <button type="button" class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
        <button type="submit" class="btn btn-primary">Update Metadata</button>
    </div>
</form>
`
	t := template.Must(template.New("edit-metadata").Parse(tmpl))
	if err := t.Execute(w, metadata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	path := r.FormValue("path")
	value := r.FormValue("value")

	updateReq := domain.UpdateMetadataRequest{}
	if path != "" {
		updateReq.Path = &path
	}
	if value != "" {
		updateReq.Value = &value
	}

	if _, err := h.service.UpdateMetadata(id, updateReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.ListMetadata(w, r)
}

func (h *Handler) DeleteMetadata(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	// Look up metadata by path to get its ID
	metadata, err := h.service.GetMetadataByPath(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := h.service.DeleteMetadata(metadata.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Storage handlers
func (h *Handler) ListStorage(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.service.ListBuckets(domain.BucketListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
        <h2 class="text-lg font-semibold">Storage Buckets</h2>
        <button class="btn btn-primary" hx-get="/web/storage/buckets/new" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
            </svg>
            New Bucket
        </button>
    </div>
    <table class="w-full">
        <thead>
            <tr>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Bucket Name</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr class="hover:bg-slate-50">
                <td class="px-6 py-4 border-b border-slate-100">
                    <div class="flex items-center gap-3">
                        <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"></path>
                        </svg>
                        <span class="font-medium">{{.Name}}</span>
                    </div>
                </td>
                <td class="px-6 py-4 border-b border-slate-100">
                    <button class="btn btn-secondary btn-sm" hx-get="/web/storage/buckets/{{.Name}}/objects" hx-target="#content">Browse Objects</button>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<div id="modal" class="hidden fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm items-start justify-center" onclick="if(event.target === this) this.style.display='none'">
    <div class="bg-white rounded-xl shadow-xl w-full max-w-lg mt-[10vh] h-fit overflow-hidden" onclick="event.stopPropagation()">
        <div id="modal-content"></div>
    </div>
</div>

<style>
#modal[style*="block"] { display: flex !important; }
</style>
`

	t := template.Must(template.New("storage-buckets").Parse(tmpl))
	if err := t.Execute(w, buckets); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewBucketForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">New Bucket</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<form hx-post="/web/storage/buckets" hx-target="#content" hx-on::after-request="if(event.detail.successful) document.getElementById('modal').style.display='none'">
    <div class="p-6">
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="name">Bucket Name</label>
            <input type="text" id="name" name="name" placeholder="my-bucket" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
    </div>
    <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3 bg-slate-50">
        <button type="button" class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
        <button type="submit" class="btn btn-primary">Create Bucket</button>
    </div>
</form>
`
	w.Write([]byte(tmpl))
}

func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if _, err := h.service.CreateBucket(domain.CreateBucketRequest{Name: name}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.ListStorage(w, r)
}

func (h *Handler) ListBucketObjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	prefix := r.URL.Query().Get("prefix")

	bucket, err := h.service.GetBucketByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	objects, err := h.service.ListObjects(domain.ObjectListOptions{BucketID: bucket.Name, Prefix: prefix})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
        <div class="flex items-center gap-3">
            <button class="btn btn-secondary btn-sm" hx-get="/web/storage" hx-target="#content">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
                </svg>
                Back
            </button>
            <h2 class="text-lg font-semibold">{{.Bucket.Name}}</h2>
        </div>
        <button class="btn btn-primary" hx-get="/web/storage/buckets/{{.Bucket.Name}}/objects/new" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"></path>
            </svg>
            Upload Object
        </button>
    </div>
    <div class="px-6 py-4 border-b border-slate-200">
        <div class="max-w-sm">
            <label class="block text-sm font-medium mb-1.5" for="prefix-filter">Filter by prefix</label>
            <input type="text" id="prefix-filter" name="prefix" hx-get="/web/storage/buckets/{{.Bucket.Name}}/objects" hx-params="*" hx-target="#content" hx-trigger="input changed delay:500ms" value="{{.Prefix}}" placeholder="folder/subfolder/" class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
    </div>
    <table class="w-full">
        <thead>
            <tr>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Path</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Updated At</th>
                <th class="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider text-slate-500 bg-slate-50 border-b border-slate-200">Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .Objects}}
            <tr class="hover:bg-slate-50">
                <td class="px-6 py-4 border-b border-slate-100">
                    <div class="flex items-center gap-3">
                        <svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path>
                        </svg>
                        <code class="bg-slate-100 px-2 py-0.5 rounded text-sm">{{.Path}}</code>
                    </div>
                </td>
                <td class="px-6 py-4 border-b border-slate-100 text-slate-500">{{.UpdatedAt.Format "2006-01-02 15:04:05"}}</td>
                <td class="px-6 py-4 border-b border-slate-100">
                    <button class="btn btn-secondary btn-sm" hx-get="/web/storage/buckets/{{$.Bucket.Name}}/objects/{{.ID}}" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">View</button>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<div id="modal" class="hidden fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm items-start justify-center" onclick="if(event.target === this) this.style.display='none'">
    <div class="bg-white rounded-xl shadow-xl w-full max-w-lg mt-[10vh] h-fit overflow-hidden" onclick="event.stopPropagation()">
        <div id="modal-content"></div>
    </div>
</div>

<style>
#modal[style*="block"] { display: flex !important; }
</style>
`

	data := struct {
		Bucket  *domain.Bucket
		Objects []*domain.Object
		Prefix  string
	}{
		Bucket:  bucket,
		Objects: objects,
		Prefix:  prefix,
	}

	t := template.Must(template.New("bucket-objects").Parse(tmpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewObjectForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	bucket, err := h.service.GetBucketByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">Upload Object to {{.Name}}</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<form hx-post="/web/storage/buckets/{{.ID}}/objects" hx-target="#content" hx-on::after-request="if(event.detail.successful) document.getElementById('modal').style.display='none'">
    <div class="p-6">
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="path">Object Path</label>
            <input type="text" id="path" name="path" placeholder="folder/file.txt" required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all">
        </div>
        <div class="mb-5">
            <label class="block text-sm font-medium mb-1.5" for="content_raw">Content</label>
            <textarea id="content_raw" name="content_raw" rows="8" placeholder="Enter file content..." required class="w-full px-3.5 py-2.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-[#2878B5] focus:ring-2 focus:ring-[#2878B5]/10 transition-all resize-none font-mono"></textarea>
        </div>
        <p class="text-xs text-slate-500">The content will be base64-encoded and stored.</p>
    </div>
    <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3 bg-slate-50">
        <button type="button" class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
        <button type="submit" class="btn btn-primary">Upload Object</button>
    </div>
</form>
`
	t := template.Must(template.New("new-object").Parse(tmpl))
	if err := t.Execute(w, bucket); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) CreateObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["name"]
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path := r.FormValue("path")
	raw := r.FormValue("content_raw")
	enc := base64.StdEncoding.EncodeToString([]byte(raw))
	if _, err := h.service.CreateObject(domain.CreateObjectRequest{BucketID: bucketName, Path: path, Content: enc}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.ListBucketObjects(w, r)
}

func (h *Handler) ViewObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["name"]
	objID := vars["objid"]
	obj, err := h.service.GetObject(objID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if obj.BucketID != bucketName {
		http.Error(w, "object not found in bucket", http.StatusNotFound)
		return
	}
	data, err := base64.StdEncoding.DecodeString(obj.Content)
	if err != nil {
		http.Error(w, "failed to decode object content", http.StatusInternalServerError)
		return
	}
	tmpl := `
<div class="px-6 py-5 border-b border-slate-200 flex justify-between items-center">
    <h3 class="text-lg font-semibold">{{.Path}}</h3>
    <button class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-all" onclick="document.getElementById('modal').style.display='none'">&times;</button>
</div>
<div class="p-6">
    <div class="flex gap-6 mb-4">
        <div>
            <span class="block text-xs uppercase tracking-wider text-slate-500 mb-1">Size</span>
            <span class="font-medium">{{.Size}} bytes</span>
        </div>
    </div>
    <div>
        <span class="block text-xs uppercase tracking-wider text-slate-500 mb-2">Content</span>
        <pre class="bg-slate-50 border border-slate-200 rounded-lg p-4 overflow-auto max-h-[50vh] text-sm font-mono whitespace-pre-wrap break-words">{{.Content}}</pre>
    </div>
</div>
<div class="px-6 py-4 border-t border-slate-200 flex justify-end bg-slate-50">
    <button class="btn btn-secondary" onclick="document.getElementById('modal').style.display='none'">Close</button>
</div>
`
	t := template.Must(template.New("view-object").Parse(tmpl))
	payload := struct {
		Path    string
		Size    int
		Content string
	}{
		Path:    obj.Path,
		Size:    len(data),
		Content: string(data),
	}
	if err := t.Execute(w, payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

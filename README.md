# GoDoIt

A simple to-do app built in Go with a UI powered by HTMX.

## About

The aim of this project is to get a better understanding of Go's default webserver capabilities before/instead of diving into a framework. As such, it is being built with a minimal dependence on third party libraries.

In the end, it has become a Go+HTMX implementation of [TodoMVC](https://todomvc.com/). The logic is ugly but functional, and I learned a lot about templating HTML in Go and how to structure api calls with HTMX.

## Skills Learned
- [x] Routing in vanilla Go. I feel that, for me, the abstractions provided by routing frameworks (e.g. router.Get() instead of router.Handle with switch statements on request.Method) are worth adding a dependency for.
- [x] HTML templating in Go. I would like to try a templating framework (e.g. [templ](https://templ.guide/)) to see if they can make things simpler, but without having seen what I might be missing I see no issues with using the built in templating for larger projects in the future.
- [x] Simple CRUD with HTMX.
- [x] Triggering HTMX calls with headers returned by previous calls.
- [x] Properly syncing/maintaining state of checkboxes (this was only an issue due to my lacking html knowledge but caused me such a headache).
- [x] Server-only state. One of the side-effects I see of HTMX is forcing developers to find the best/most efficient patterns for storing state on the server-side only and sending only the data necessary for each update, which I think is a very good thing.

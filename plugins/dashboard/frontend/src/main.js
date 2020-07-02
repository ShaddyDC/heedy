import vuexModule from "./main/vuex.js";
import Injector from "./main/injector";

function setup(frontend) {
  frontend.store.registerModule("dashboard", vuexModule);
  frontend.inject("dashboard", new Injector(frontend));

  frontend.objects.setType({
    type: "dashboard",
    title: "Dashboard",
    list_title: "Dashboards",
    icon: "dashboard",
  });

  if (frontend.info.user != null) {
    frontend.objects.addCreator({
      key: "dashboard",
      title: "Dashboard",
      description: "Display data from multiple sources",
      icon: "dashboard",
      fn: async () => {
        let res = await frontend.rest("POST", "/api/objects", {
          name: "My Dashboard",
          type: "dashboard",
        });
        if (res.response.ok) {
          frontend.router.push({ path: `/objects/${res.data.id}/update` });
        } else {
          frontend.store.dispatch("errnotify", res.data);
        }
      },
    });
  }
}

export default setup;

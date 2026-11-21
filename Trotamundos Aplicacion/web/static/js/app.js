(function () {
  var activeLeafletMap = null;

  function closeModal() {
    var root = document.getElementById("modal-root");
    if (root) {
      root.innerHTML = "";
    }
  }

  function switchReservationFields(type) {
    var groups = document.querySelectorAll("[data-res-fields]");
    groups.forEach(function (el) {
      if (el.getAttribute("data-res-fields") === type) {
        el.classList.remove("hidden");
      } else {
        el.classList.add("hidden");
      }
    });
  }

  function dismissToast(el) {
    if (!el) return;
    var card = el.closest ? el.closest(".toast-item") : null;
    if (card) {
      card.remove();
    }
  }

  function fillPNR(code) {
    var input = document.getElementById("pnr-code-input");
    if (input) {
      input.value = code;
      input.focus();
    }
  }

  function fillExternalEmail(preset) {
    var presets = {
      booking: {
        from: "reservations@booking.com",
        subject: "Confirmación de alojamiento en Tokio — Hotel Gracery Shinjuku",
        body: "Tu estadía está confirmada. Código de confirmación: BKG-7741. Check-in 15:00 en Shinjuku, Tokio."
      },
      aerolineas: {
        from: "confirmaciones@aerolineas.com.ar",
        subject: "Pasaje emitido Buenos Aires → Bariloche (AR 1680)",
        body: "Reserva confirmada con Aerolíneas Argentinas. Localizador PNR: AR-1140. Salida desde Aeroparque (AEP) hacia Bariloche (BRC)."
      },
      shinkansen: {
        from: "reservas@jrpass.com",
        subject: "Boleto emitido Tokaido Shinkansen Nozomi 21",
        body: "Tu asiento está reservado. Código PNR: JR-9914. Trayecto Tokio Central → Kioto Central."
      },
      untrusted: {
        from: "ofertas@agencia-desconocida.xyz",
        subject: "Confirmación promocional de viaje",
        body: "Código PNR: EXT-9999. Correo de un dominio que no está en tu lista blanca de remitentes confiables."
      }
    };
    var data = presets[preset];
    if (!data) return;
    var fromEl = document.getElementById("ext-email-from");
    var subjEl = document.getElementById("ext-email-subject");
    var bodyEl = document.getElementById("ext-email-body");
    if (fromEl) fromEl.value = data.from;
    if (subjEl) subjEl.value = data.subject;
    if (bodyEl) bodyEl.value = data.body;
  }

  function closeExploreMenu() {
    var wrapper = document.getElementById("explore-dropdown-wrapper");
    if (wrapper) {
      wrapper.removeAttribute("open");
    }
  }

  function toggleExploreMenu(evt) {
    if (evt) {
      evt.stopPropagation();
    }
    var wrapper = document.getElementById("explore-dropdown-wrapper");
    if (!wrapper) return;
    if (wrapper.hasAttribute("open")) {
      wrapper.removeAttribute("open");
    } else {
      wrapper.setAttribute("open", "");
    }
  }

  function initCompanionMap() {
    var container = document.getElementById("leaflet-companion-map");
    if (!container || typeof window.L === "undefined") {
      return;
    }

    var originLat = parseFloat(container.getAttribute("data-origin-lat") || "0");
    var originLon = parseFloat(container.getAttribute("data-origin-lon") || "0");
    var originLabel = container.getAttribute("data-origin-label") || "Punto de llegada";
    var hotelLat = parseFloat(container.getAttribute("data-hotel-lat") || "0");
    var hotelLon = parseFloat(container.getAttribute("data-hotel-lon") || "0");
    var hotelName = container.getAttribute("data-hotel-name") || "Alojamiento";
    var rawPolyline = container.getAttribute("data-polyline") || "[]";

    if (isNaN(originLat) || isNaN(originLon) || isNaN(hotelLat) || isNaN(hotelLon)) {
      return;
    }
    if (originLat === 0 && originLon === 0 && hotelLat === 0 && hotelLon === 0) {
      return;
    }

    if (activeLeafletMap) {
      try {
        activeLeafletMap.remove();
      } catch (e) {
        // ignore cleanup errors on detached DOM
      }
      activeLeafletMap = null;
    }
    if (container._leaflet_id) {
      container._leaflet_id = null;
    }

    var centerLat = (originLat + hotelLat) / 2;
    var centerLon = (originLon + hotelLon) / 2;

    var map = L.map(container, {
      scrollWheelZoom: false,
      zoomControl: true
    }).setView([centerLat, centerLon], 12);

    activeLeafletMap = map;

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
    }).addTo(map);

    var originIcon = L.divIcon({
      className: "custom-leaflet-marker",
      html: '<div style="width:28px;height:28px;border-radius:9999px;background:#0f172a;color:#ffffff;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:12px;border:3px solid #ffffff;box-shadow:0 4px 10px rgba(15,23,42,0.35);">A</div>',
      iconSize: [28, 28],
      iconAnchor: [14, 14],
      popupAnchor: [0, -14]
    });

    var hotelIcon = L.divIcon({
      className: "custom-leaflet-marker",
      html: '<div style="width:32px;height:32px;border-radius:9999px;background:#e11d48;color:#ffffff;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:12px;border:3px solid #ffffff;box-shadow:0 4px 12px rgba(225,29,72,0.45);">H</div>',
      iconSize: [32, 32],
      iconAnchor: [16, 16],
      popupAnchor: [0, -16]
    });

    var originMarker = L.marker([originLat, originLon], { icon: originIcon })
      .addTo(map)
      .bindPopup("<strong>Llegada:</strong> " + originLabel);

    var hotelMarker = L.marker([hotelLat, hotelLon], { icon: hotelIcon })
      .addTo(map)
      .bindPopup("<strong>Destino:</strong> " + hotelName);

    var latLngs = [];
    try {
      var parsed = JSON.parse(rawPolyline);
      if (Array.isArray(parsed)) {
        parsed.forEach(function (pt) {
          if (Array.isArray(pt) && pt.length >= 2) {
            latLngs.push([parseFloat(pt[0]), parseFloat(pt[1])]);
          } else if (pt && typeof pt === "object") {
            var lat = pt.lat !== undefined ? pt.lat : pt.Lat;
            var lon = pt.lon !== undefined ? pt.lon : pt.Lon;
            if (lat !== undefined && lon !== undefined) {
              latLngs.push([parseFloat(lat), parseFloat(lon)]);
            }
          }
        });
      }
    } catch (err) {
      latLngs = [];
    }

    if (latLngs.length < 2) {
      latLngs = [
        [originLat, originLon],
        [hotelLat, hotelLon]
      ];
    }

    // Subtle outer halo + primary brand crimson line
    L.polyline(latLngs, {
      color: "#fff1f2",
      weight: 8,
      opacity: 0.85,
      lineCap: "round",
      lineJoin: "round"
    }).addTo(map);

    var routeLine = L.polyline(latLngs, {
      color: "#e11d48",
      weight: 4.5,
      opacity: 0.95,
      lineCap: "round",
      lineJoin: "round"
    }).addTo(map);

    var group = L.featureGroup([originMarker, hotelMarker, routeLine]);
    map.fitBounds(group.getBounds(), { padding: [40, 40], maxZoom: 15 });

    setTimeout(function () {
      map.invalidateSize();
    }, 120);
  }

  window.closeModal = closeModal;
  window.initCompanionMap = initCompanionMap;
  window.TrotamundosUI = {
    closeModal: closeModal,
    switchReservationFields: switchReservationFields,
    dismissToast: dismissToast,
    fillPNR: fillPNR,
    fillExternalEmail: fillExternalEmail,
    initCompanionMap: initCompanionMap,
    toggleExploreMenu: toggleExploreMenu,
    closeExploreMenu: closeExploreMenu
  };

  document.body.addEventListener("closeModal", closeModal);

  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    if (!evt.detail || !evt.detail.xhr) return;
    var status = evt.detail.xhr.status;
    if (status >= 400 && status < 500) {
      evt.detail.shouldSwap = true;
      evt.detail.isError = false;
    }
  });

  document.body.addEventListener("htmx:afterSwap", function () {
    initCompanionMap();
  });

  document.addEventListener("DOMContentLoaded", function () {
    initCompanionMap();
  });

  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape") {
      closeModal();
      closeExploreMenu();
    }
  });

  document.addEventListener("click", function (e) {
    var wrapper = document.getElementById("explore-dropdown-wrapper");
    if (wrapper && !wrapper.contains(e.target)) {
      closeExploreMenu();
    }
  });

  // Auto-dismiss HTMX toasts after 5.2 seconds
  var observer = new MutationObserver(function (mutations) {
    mutations.forEach(function (m) {
      m.addedNodes.forEach(function (node) {
        if (node.nodeType === 1) {
          var toasts = [];
          if (node.classList && node.classList.contains("toast-item")) {
            toasts.push(node);
          } else if (node.querySelectorAll) {
            node.querySelectorAll(".toast-item").forEach(function (t) {
              toasts.push(t);
            });
          }
          toasts.forEach(function (toastEl) {
            setTimeout(function () {
              if (toastEl && toastEl.parentNode) {
                toastEl.remove();
              }
            }, 5200);
          });
        }
      });
    });
  });

  var toastContainer = document.getElementById("toast-container");
  if (toastContainer) {
    observer.observe(toastContainer, { childList: true, subtree: true });
  }
})();

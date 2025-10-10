import {Control, type ControlOptions, DomUtil, Map as LeafletMap} from 'leaflet';

export class CustomControl extends Control {
    private el: HTMLDivElement | undefined;

    constructor(options?: ControlOptions) {
        super(options);
    }

    public setText(text: string) {
        if (this.el === undefined) {
            return;
        }
        this.el.innerText = text;
    }

    public override onAdd(map: LeafletMap): HTMLElement {
        this.el = DomUtil.create('div', 'leaflet-custom-control');
        return this.el;
    }

    public override onRemove(_map: LeafletMap): void {
        // do nothing
    }
}

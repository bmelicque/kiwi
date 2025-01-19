/**
 * This file serves documentation purposes.
 * It includes all utility needed for running a program.
 * This is the source of the minified code injected at transpiling.
 */

/**
 * The base class for sum types (tagged unions)
 */
export class Sum {
	constructor(tag, value) {
		this.tag = tag;
		this.value = value;
	}
}

/**
 * Wraps native methods related to DOM. Handles pointers, this values, etc.
 *
 * @param {Node} object The object that holds the method
 * @param {string} method The name of the method to bind/call
 * @param {boolean} returnsPointer
 * @returns
 */
export function wrapNodeMethod(object, method, returnsPointer) {
	// readonly properties are represented as getters
	if (typeof object[method] != "function") {
		if (returnsPointer) {
			return () => new NodePointer(object[method]);
		} else {
			return () => object[method];
		}
	}

	const wrapped = (...args) =>
		object[method].apply(
			object,
			args.map((arg) => (arg instanceof NodePointer ? arg.get() : arg))
		);

	if (returnsPointer) {
		return (...a) => new NodePointer(wrapped(...a));
	} else {
		return wrapped;
	}
}

export function bind(object, method) {
	return object[method].bind(object);
}

export class Pointer {
	constructor(ctx, name) {
		this.ctx = ctx;
		this.name = name;
	}

	get() {
		return this.ctx ? this.ctx[this.name] : this.name;
	}

	set(value) {
		this.ctx ? (this.ctx[this.name] = value) : (this.name = value);
	}
}

/**
 * A pointer to a Node. Doesn't need context like regular pointer,
 * but needs to replace Node its owner document in case of update.
 */
export class NodePointer {
	constructor(value) {
		this.value = value;
	}

	get() {
		return this.value;
	}

	set(value) {
		this.value.parentNode?.replaceChild(this.value, value);
		this.value = value;
	}
}

/**
 * Deep comparison between two objects
 */
export function equals(a, b) {
	if (typeof a !== typeof b) return false;
	if (typeof a !== "object" || a == null || b == null) return a === b;
	if (a.constructor !== b.constructor) return false;
	if (a instanceof NodePointer) return a.get() == b.get();
	if (Array.isArray(a) && a.length !== b.length) return false;
	return !Object.keys(a).find((k) => !equals(a[k], b[k]));
}

export function getDocument() {
	return new NodePointer(document);
}

/**
 * Wrapper around document.createElement to handle some shorthands
 * @param {string} string
 */
export function createElement(string) {
	const match = string.match(/^(\w[\w\-_]*)?(?:#(\w[\w\-_]*))?((?:\.\w[\w\-_]*)*)$/);
	if (!match[0]) throw new Error("Invalid selector");
	const el = document.createElement(match[1] || "div");
	if (match[2]) el.id = match[2].slice(1);
	if (match[3]) el.classList.add(...match[3].split(".").slice(1));
	return el;
}

export class DocumentBody extends Sum {}

// TODO: use actual Option type
export function getDocumentBody(document) {
	if (document instanceof NodePointer) document = document.get();
	if (document == null) return null;

	const body = document.body;
	const tag = body instanceof HTMLBodyElement ? "Body" : "Frame";
	return new DocumentBody(tag, body);
}

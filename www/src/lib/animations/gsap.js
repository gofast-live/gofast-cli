import gsap from 'gsap';
import { TextPlugin } from 'gsap/TextPlugin';
import { ScrollToPlugin } from 'gsap/ScrollToPlugin';

if (typeof window !== 'undefined') {
	gsap.registerPlugin(TextPlugin, ScrollToPlugin);
}

export { gsap };

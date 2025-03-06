/**
 * Helpers to create HTML elements used by Choices
 * Can be overridden by providing `callbackOnCreateTemplates` option.
 * `Choices.defaults.templates` allows access to the default template methods from `callbackOnCreateTemplates`
 */
import { Templates as TemplatesInterface } from './interfaces/templates';
declare const templates: TemplatesInterface;
export default templates;

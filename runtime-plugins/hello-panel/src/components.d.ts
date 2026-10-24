// Vuetify components are registered globally by the host app. Declaring them
// here keeps vue-tsc happy without shipping a second copy of Vuetify.
declare module 'vue' {
  export interface GlobalComponents {
    VAlert: typeof import('vuetify/components')['VAlert']
    VBtn: typeof import('vuetify/components')['VBtn']
    VCard: typeof import('vuetify/components')['VCard']
    VCardText: typeof import('vuetify/components')['VCardText']
    VCardTitle: typeof import('vuetify/components')['VCardTitle']
    VChip: typeof import('vuetify/components')['VChip']
    VDivider: typeof import('vuetify/components')['VDivider']
    VIcon: typeof import('vuetify/components')['VIcon']
  }
}

export {}

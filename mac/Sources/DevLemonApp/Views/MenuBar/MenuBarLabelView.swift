import SwiftUI

public struct MenuBarLabelView: View {
    @ObservedObject var provider = MenuBarImageProvider.shared

    public init() {}

    public var body: some View {
        ZStack {
            if let image = provider.currentImage {
                Image(nsImage: image)
            } else {
                MonochromeLemonIcon(size: 14)
            }
        }
        .onAppear {
            provider.regenerateImage()
        }
    }
}

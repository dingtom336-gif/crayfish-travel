export default function PrivacyPage() {
  return (
    <main className="bg-gray-50 py-12 flex-1">
      <div className="max-w-3xl mx-auto px-4">
        <h1 className="text-2xl font-bold mb-8">隐私政策</h1>
        <div className="bg-white rounded-xl p-8 shadow-sm space-y-6 text-sm text-gray-600 leading-relaxed">
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">1. 信息收集</h2>
            <p>我们仅收集提供服务所必需的最少信息，包括旅行需求（目的地、日期、预算、人数）及支付阶段所需的实名信息。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">2. 信息保护</h2>
            <p>所有敏感个人信息采用AES-256-GCM加密存储，并在72小时后自动删除。我们不会将您的个人信息出售或分享给无关第三方。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">3. 信息使用</h2>
            <p>收集的信息仅用于：匹配旅行方案、处理订单和支付、提供售后服务。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">4. 用户权利</h2>
            <p>您有权随时要求删除您的个人信息。联系客服即可发起删除请求，我们将在24小时内处理。</p>
          </section>
        </div>
      </div>
    </main>
  )
}
